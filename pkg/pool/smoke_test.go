package pool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSmokeWritesSummaryAndParsesMutationReport(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "manifest.json")
	writeManifest(t, manifestPath, []Repo{{
		Name:   "cobra",
		URL:    "https://example.com/cobra.git",
		Target: "./doc",
		Lane:   "tuning",
		Domain: "cli",
	}})

	runner := &fakeRunner{run: func(spec CommandSpec) (CommandResult, error) {
		switch spec.Path {
		case "git":
			dest := spec.Args[len(spec.Args)-1]
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return CommandResult{}, err
			}
		case "cervomut":
			out := flagValue(spec.Args, "--out")
			if strings.Contains(strings.Join(spec.Args, " "), "ci-balanced") {
				reportPath := filepath.Join(out, "mutation-report.json")
				if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
					return CommandResult{}, err
				}
				if err := os.WriteFile(reportPath, []byte(`{"summary":{"total":10,"killed":7,"survived":2,"not_covered":1,"timed_out":0,"compile_error":0,"skipped":0,"score":77.7,"test_efficacy":77.7}}`), 0o644); err != nil {
					return CommandResult{}, err
				}
			}
		}
		return CommandResult{ExitCode: 0}, nil
	}}

	run, err := RunSmoke(context.Background(), SmokeOptions{
		ManifestPath:           manifestPath,
		WorkRoot:               filepath.Join(root, "work"),
		Names:                  []string{"cobra"},
		RunMutation:            true,
		MaxMutants:             10,
		Workers:                2,
		CloneTimeoutSeconds:    60,
		TestTimeoutSeconds:     60,
		DryRunTimeoutSeconds:   60,
		MutationTimeoutSeconds: 60,
		CervoBinary:            "cervomut",
		GitBinary:              "git",
		Runner:                 runner,
	})
	if err != nil {
		t.Fatalf("RunSmoke returned error: %v", err)
	}
	if len(run.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(run.Results))
	}
	got := run.Results[0]
	if got.Clone != "ok" {
		t.Fatalf("clone = %q, want ok", got.Clone)
	}
	if got.Mutants == nil || *got.Mutants != 10 || got.Killed == nil || *got.Killed != 7 || got.Survived == nil || *got.Survived != 2 || got.NotCovered == nil || *got.NotCovered != 1 {
		t.Fatalf("mutation metrics not parsed: %+v", got)
	}
	if _, err := os.Stat(run.SummaryPath); err != nil {
		t.Fatalf("summary missing: %v", err)
	}
	if len(runner.specs) != 4 {
		t.Fatalf("command count = %d, want 4", len(runner.specs))
	}
}

type fakeRunner struct {
	specs []CommandSpec
	run   func(CommandSpec) (CommandResult, error)
}

func (f *fakeRunner) Run(_ context.Context, spec CommandSpec) (CommandResult, error) {
	f.specs = append(f.specs, spec)
	return f.run(spec)
}

func writeManifest(t *testing.T, path string, repos []Repo) {
	t.Helper()
	data, err := json.Marshal(Manifest{SchemaVersion: "1", Repos: repos})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func flagValue(args []string, flag string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}
