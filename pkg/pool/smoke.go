package pool

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cervantesh/CervoMutants/pkg/extcompare"
)

type SmokeOptions struct {
	ManifestPath           string
	WorkRoot               string
	Names                  []string
	Limit                  int
	RunMutation            bool
	MaxMutants             int
	Workers                int
	CloneTimeoutSeconds    int
	TestTimeoutSeconds     int
	DryRunTimeoutSeconds   int
	MutationTimeoutSeconds int
	CervoBinary            string
	GitBinary              string
	Runner                 CommandRunner
}

type SmokeResult struct {
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Target          string   `json:"target"`
	Lane            string   `json:"lane"`
	Domain          string   `json:"domain"`
	Clone           string   `json:"clone"`
	BaselineExit    *int     `json:"baseline_exit"`
	BaselineSeconds float64  `json:"baseline_seconds"`
	DryRunExit      *int     `json:"dry_run_exit"`
	DryRunSeconds   float64  `json:"dry_run_seconds"`
	MutationExit    *int     `json:"mutation_exit"`
	MutationSeconds float64  `json:"mutation_seconds"`
	Mutants         *int     `json:"mutants"`
	Killed          *int     `json:"killed"`
	Survived        *int     `json:"survived"`
	NotCovered      *int     `json:"not_covered"`
	Score           *float64 `json:"score"`
	Notes           string   `json:"notes"`
	ElapsedSeconds  float64  `json:"elapsed_seconds"`
}

func RunSmoke(ctx context.Context, opts SmokeOptions) (RunSummary[SmokeResult], error) {
	manifest, err := LoadManifest(opts.ManifestPath)
	if err != nil {
		return RunSummary[SmokeResult]{}, err
	}
	repos := FilterRepos(manifest, opts.Names, opts.Limit)
	if err := os.MkdirAll(opts.WorkRoot, 0o755); err != nil {
		return RunSummary[SmokeResult]{}, err
	}
	runner := opts.Runner
	if runner == nil {
		runner = RealCommandRunner{}
	}
	gitBinary, err := requiredBinary("git", defaultPath(opts.GitBinary, "git"))
	if err != nil {
		return RunSummary[SmokeResult]{}, err
	}
	cervoBinary, err := requiredBinary("cervomut", opts.CervoBinary)
	if err != nil {
		return RunSummary[SmokeResult]{}, err
	}
	results := make([]SmokeResult, 0, len(repos))
	for _, repo := range repos {
		started := time.Now()
		result := SmokeResult{
			Name:   repo.Name,
			URL:    repo.URL,
			Target: repo.Target,
			Lane:   repo.Lane,
			Domain: repo.Domain,
			Clone:  "pending",
		}
		repoDir := filepath.Join(opts.WorkRoot, repo.Name)
		if _, statErr := os.Stat(repoDir); statErr != nil {
			cloneExit, runErr := runSimpleCommand(ctx, runner, CommandSpec{
				Path:    gitBinary,
				Args:    []string{"clone", "--depth", "1", repo.URL, repoDir},
				Dir:     opts.WorkRoot,
				LogPath: filepath.Join(opts.WorkRoot, repo.Name+"-clone.log"),
				Timeout: time.Duration(opts.CloneTimeoutSeconds) * time.Second,
			})
			if runErr != nil {
				return RunSummary[SmokeResult]{}, runErr
			}
			if cloneExit != 0 {
				result.Clone = "failed"
				result.Notes = "clone exit " + strconv.Itoa(cloneExit)
				result.ElapsedSeconds = roundSeconds(started)
				results = append(results, result)
				continue
			}
		}
		result.Clone = "ok"

		baselineOut := filepath.Join(opts.WorkRoot, repo.Name+"-baseline.log")
		baselineStart := time.Now()
		baselineExit, runErr := runSimpleCommand(ctx, runner, CommandSpec{
			Path:    "go",
			Args:    []string{"test", repo.Target},
			Dir:     repoDir,
			LogPath: baselineOut,
			Timeout: time.Duration(opts.TestTimeoutSeconds) * time.Second,
		})
		if runErr != nil {
			return RunSummary[SmokeResult]{}, runErr
		}
		result.BaselineExit = intPtr(baselineExit)
		result.BaselineSeconds = roundSeconds(baselineStart)

		outRoot := filepath.Join(opts.WorkRoot, "reports", repo.Name)
		dryRunLog := filepath.Join(opts.WorkRoot, repo.Name+"-dry-run.log")
		dryRunStart := time.Now()
		dryRunExit, runErr := runSimpleCommand(ctx, runner, CommandSpec{
			Path:    cervoBinary,
			Args:    []string{"run", repo.Target, "--dry-run", "--policy", "ci-fast", "--max-mutants", strconv.Itoa(opts.MaxMutants), "--workers", strconv.Itoa(opts.Workers), "--out", outRoot},
			Dir:     repoDir,
			LogPath: dryRunLog,
			Timeout: time.Duration(opts.DryRunTimeoutSeconds) * time.Second,
		})
		if runErr != nil {
			return RunSummary[SmokeResult]{}, runErr
		}
		result.DryRunExit = intPtr(dryRunExit)
		result.DryRunSeconds = roundSeconds(dryRunStart)

		if opts.RunMutation {
			mutationLog := filepath.Join(opts.WorkRoot, repo.Name+"-mutation.log")
			mutationStart := time.Now()
			mutationExit, runErr := runSimpleCommand(ctx, runner, CommandSpec{
				Path:    cervoBinary,
				Args:    []string{"run", repo.Target, "--policy", "ci-balanced", "--max-mutants", strconv.Itoa(opts.MaxMutants), "--workers", strconv.Itoa(opts.Workers), "--out", outRoot},
				Dir:     repoDir,
				LogPath: mutationLog,
				Timeout: time.Duration(opts.MutationTimeoutSeconds) * time.Second,
			})
			if runErr != nil {
				return RunSummary[SmokeResult]{}, runErr
			}
			result.MutationExit = intPtr(mutationExit)
			result.MutationSeconds = roundSeconds(mutationStart)
			reportPath := filepath.Join(outRoot, "mutation-report.json")
			if _, statErr := os.Stat(reportPath); statErr == nil {
				parsed, parseErr := extcompare.ParseCervo(reportPath)
				if parseErr != nil {
					return RunSummary[SmokeResult]{}, parseErr
				}
				result.Mutants = intPtr(parsed.Total)
				result.Killed = intPtr(parsed.Killed)
				result.Survived = intPtr(parsed.Survived)
				result.NotCovered = intPtr(parsed.NotCovered)
				result.Score = floatPtr(parsed.Score)
			}
		}
		result.ElapsedSeconds = roundSeconds(started)
		results = append(results, result)
	}
	path := summaryPath(opts.WorkRoot)
	if err := writeJSON(path, results); err != nil {
		return RunSummary[SmokeResult]{}, err
	}
	return RunSummary[SmokeResult]{Results: results, SummaryPath: path}, nil
}

func runSimpleCommand(ctx context.Context, runner CommandRunner, spec CommandSpec) (int, error) {
	result, err := runner.Run(ctx, spec)
	if err != nil {
		return 0, err
	}
	return result.ExitCode, nil
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func roundSeconds(start time.Time) float64 {
	return float64(time.Since(start).Milliseconds()) / 1000
}
