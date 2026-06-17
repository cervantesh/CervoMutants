package pool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cervantesh/cervo-mutants/internal/testharness"
)

func TestRunBenchmarkWritesSummaryEvaluatesThresholdsAndUsesPartialFallback(t *testing.T) {
	fixture := testharness.NewDir(t)
	corpusPath := fixture.WriteJSON(t, "corpus.json", BenchmarkCorpus{
		SchemaVersion: "1",
		TrackingIssue: "#80",
		Entries: []BenchmarkEntry{{
			Name:                   "cobra-doc",
			URL:                    "https://example.com/cobra.git",
			Ref:                    "v1.10.1",
			Target:                 "./doc",
			Size:                   "small",
			ResourceClass:          "low",
			Policy:                 "comparison-safe",
			MaxMutants:             12,
			Workers:                3,
			CloneTimeoutSeconds:    30,
			BaselineTimeoutSeconds: 30,
			DryRunTimeoutSeconds:   30,
			MutationTimeoutSeconds: 30,
			Thresholds: BenchmarkThresholds{
				MaxBaselineSeconds:  30,
				MaxDryRunSeconds:    30,
				MaxMutationSeconds:  30,
				MaxPeakMemoryMB:     8,
				MinMutantsPerSecond: 0.1,
				MinExecutedMutants:  10,
			},
		}},
	})

	runner := &fakeRunner{run: func(spec CommandSpec) (CommandResult, error) {
		switch spec.Path {
		case "git":
			if len(spec.Args) > 0 && spec.Args[0] == "clone" {
				dest := spec.Args[len(spec.Args)-1]
				if err := os.MkdirAll(dest, 0o755); err != nil {
					return CommandResult{}, err
				}
			}
		case "cervomut":
			if strings.Contains(strings.Join(spec.Args, " "), "--dry-run") {
				return CommandResult{ExitCode: 0}, nil
			}
			out := flagValue(spec.Args, "--out")
			reportPath := filepath.Join(out, "partial-mutation-report.json")
			if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
				return CommandResult{}, err
			}
			report := map[string]any{
				"summary": map[string]any{
					"total":             12,
					"generated_mutants": 12,
					"executed_mutants":  10,
					"effective_mutants": 8,
					"score_denominator": 10,
					"killed":            6,
					"survived":          2,
					"timed_out":         1,
					"compile_error":     1,
				},
				"mutants": []map[string]any{
					{"mutant_id": "m1", "status": "killed", "memory_peak_bytes": 2 * 1024 * 1024},
					{"mutant_id": "m2", "status": "survived", "memory_peak_bytes": 5 * 1024 * 1024},
				},
			}
			data, err := json.Marshal(report)
			if err != nil {
				return CommandResult{}, err
			}
			if err := os.WriteFile(reportPath, data, 0o644); err != nil {
				return CommandResult{}, err
			}
		}
		return CommandResult{ExitCode: 0}, nil
	}}

	run, err := RunBenchmark(context.Background(), BenchmarkOptions{
		CorpusPath:  corpusPath,
		WorkRoot:    fixture.Path("work"),
		OutputRoot:  fixture.Path("out"),
		Names:       []string{"cobra-doc"},
		CervoBinary: "cervomut",
		GitBinary:   "git",
		Runner:      runner,
	})
	if err != nil {
		t.Fatalf("RunBenchmark returned error: %v", err)
	}
	if len(run.Results) != 1 {
		t.Fatalf("results = %d, want 1", len(run.Results))
	}
	got := run.Results[0]
	if got.Status != "pass" || got.Clone != "ok" || got.Checkout != "ok" || !got.PartialReportUsed {
		t.Fatalf("unexpected benchmark result: %+v", got)
	}
	if got.Metrics.ExecutedMutants != 10 || got.Metrics.EffectiveMutants != 8 || got.Metrics.ScoreDenominator != 10 {
		t.Fatalf("benchmark metrics mismatch: %+v", got.Metrics)
	}
	if got.Metrics.MaxPeakMemoryMB != 5 || got.Metrics.MutantsPerSecond <= 0 {
		t.Fatalf("benchmark performance metrics mismatch: %+v", got.Metrics)
	}
	if benchmarkHasFailedCheck(got.Checks) {
		t.Fatalf("benchmark checks should pass: %+v", got.Checks)
	}
	if len(runner.specs) != 6 {
		t.Fatalf("command count = %d, want 6", len(runner.specs))
	}
	if _, err := os.Stat(run.SummaryPath); err != nil {
		t.Fatalf("summary missing: %v", err)
	}

	var summary BenchmarkSummaryFile
	data, err := os.ReadFile(run.SummaryPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("summary JSON invalid: %v\n%s", err, data)
	}
	if summary.Totals.Passed != 1 || summary.Totals.Failed != 0 || summary.Totals.Errored != 0 {
		t.Fatalf("summary totals mismatch: %+v", summary.Totals)
	}
}

func TestRunBenchmarkFailsThresholdsAndResumeSkipsExistingResults(t *testing.T) {
	fixture := testharness.NewDir(t)
	corpusPath := fixture.WriteJSON(t, "corpus.json", BenchmarkCorpus{
		SchemaVersion: "1",
		Entries: []BenchmarkEntry{{
			Name:       "logrus",
			URL:        "https://example.com/logrus.git",
			Target:     "./...",
			Policy:     "comparison-safe",
			MaxMutants: 5,
			Thresholds: BenchmarkThresholds{
				MaxMutationSeconds: 0.0001,
				MinExecutedMutants: 20,
			},
		}},
	})

	runner := &fakeRunner{run: func(spec CommandSpec) (CommandResult, error) {
		switch spec.Path {
		case "git":
			if len(spec.Args) > 0 && spec.Args[0] == "clone" {
				dest := spec.Args[len(spec.Args)-1]
				if err := os.MkdirAll(dest, 0o755); err != nil {
					return CommandResult{}, err
				}
			}
		case "cervomut":
			if strings.Contains(strings.Join(spec.Args, " "), "--dry-run") {
				return CommandResult{ExitCode: 0}, nil
			}
			out := flagValue(spec.Args, "--out")
			reportPath := filepath.Join(out, "mutation-report.json")
			if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
				return CommandResult{}, err
			}
			report := map[string]any{
				"summary": map[string]any{
					"generated_mutants": 5,
					"executed_mutants":  5,
					"effective_mutants": 4,
					"score_denominator": 4,
					"killed":            3,
					"survived":          1,
				},
				"mutants": []map[string]any{
					{"mutant_id": "m1", "status": "killed", "memory_peak_bytes": 1024},
				},
			}
			data, err := json.Marshal(report)
			if err != nil {
				return CommandResult{}, err
			}
			if err := os.WriteFile(reportPath, data, 0o644); err != nil {
				return CommandResult{}, err
			}
		}
		return CommandResult{ExitCode: 0}, nil
	}}

	firstRun, err := RunBenchmark(context.Background(), BenchmarkOptions{
		CorpusPath:  corpusPath,
		WorkRoot:    fixture.Path("work"),
		OutputRoot:  fixture.Path("out"),
		CervoBinary: "cervomut",
		GitBinary:   "git",
		Runner:      runner,
	})
	if err == nil || !strings.Contains(err.Error(), "benchmark threshold failed") {
		t.Fatalf("RunBenchmark should fail thresholds, got %v", err)
	}
	if len(firstRun.Results) != 1 || firstRun.Results[0].Status != "fail" {
		t.Fatalf("first run results mismatch: %+v", firstRun.Results)
	}
	if len(runner.specs) != 4 {
		t.Fatalf("first run command count = %d, want 4", len(runner.specs))
	}

	resumeRunner := &fakeRunner{run: func(spec CommandSpec) (CommandResult, error) {
		t.Fatalf("resume should not execute commands: %+v", spec)
		return CommandResult{}, nil
	}}
	secondRun, err := RunBenchmark(context.Background(), BenchmarkOptions{
		CorpusPath:  corpusPath,
		WorkRoot:    fixture.Path("work"),
		OutputRoot:  fixture.Path("out"),
		Resume:      true,
		CervoBinary: "cervomut",
		GitBinary:   "git",
		Runner:      resumeRunner,
	})
	if err == nil || !strings.Contains(err.Error(), "benchmark threshold failed") {
		t.Fatalf("resume RunBenchmark should preserve threshold failure, got %v", err)
	}
	if len(secondRun.Results) != 1 {
		t.Fatalf("resume results = %d, want 1", len(secondRun.Results))
	}
	if !containsNote(secondRun.Results[0].Notes, "resumed from existing summary") {
		t.Fatalf("resume note missing: %+v", secondRun.Results[0])
	}
	if len(resumeRunner.specs) != 0 {
		t.Fatalf("resume should not run commands, ran %d", len(resumeRunner.specs))
	}
}
