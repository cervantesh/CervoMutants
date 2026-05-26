package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitea.cervbox.synology.me/CervoSoft/cervo-mutant/pkg/config"
)

func writeFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fixture\n\ngo 1.25.6\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc.go"), []byte(`package fixture

func IsPositiveOrZero(n int) bool {
	return n >= 0
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc_test.go"), []byte(`package fixture

import "testing"

func TestIsPositiveOrZero(t *testing.T) {
	if !IsPositiveOrZero(1) {
		t.Fatal("want positive")
	}
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func isolateArtifacts(cfg *config.Config, dir string) {
	cfg.Reports.Output = filepath.Join(dir, ".cervomut", "reports")
	cfg.Cache.Path = filepath.Join(dir, ".cervomut", "cache")
	cfg.Selection.CoverageProfile = filepath.Join(dir, ".cervomut", "coverage.out")
	cfg.Selection.TimingsPath = filepath.Join(dir, ".cervomut", "timings.json")
}

func TestRunDryRunDiscoversMutantsWithoutChangingWorkspace(t *testing.T) {
	dir := writeFixture(t)
	before, err := os.ReadFile(filepath.Join(dir, "calc.go"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Defaults()
	cfg.Tests.Command = []string{"go", "test", "./..."}
	isolateArtifacts(&cfg, dir)

	result, err := New(cfg).Run(context.Background(), RunRequest{Targets: []string{dir}, DryRun: true})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary.Total == 0 {
		t.Fatal("dry-run discovered no mutants")
	}
	after, err := os.ReadFile(filepath.Join(dir, "calc.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("dry-run changed source workspace")
	}
}

func TestRunClassifiesSurvivorAndWritesReports(t *testing.T) {
	dir := writeFixture(t)
	cfg := config.Defaults()
	cfg.Tests.Command = []string{"go", "test", "./..."}
	cfg.Tests.Timeout = 10_000_000_000
	isolateArtifacts(&cfg, dir)
	cfg.Execution.Workers = 1

	result, err := New(cfg).Run(context.Background(), RunRequest{Targets: []string{dir}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Summary.Total == 0 {
		t.Fatal("run discovered no mutants")
	}
	if result.Summary.Survived == 0 {
		t.Fatalf("expected weak fixture test to leave a survivor: %+v", result.Summary)
	}
	reportPath := filepath.Join(cfg.Reports.Output, "mutation-report.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("report was not written: %v", err)
	}
	if !strings.Contains(string(data), `"schema_version": "1"`) {
		t.Fatalf("report missing schema version: %s", data)
	}
}

func TestCoverageSelectionUsesCoverageProfileAndRecordsTiming(t *testing.T) {
	dir := writeFixture(t)
	cfg := config.Defaults()
	cfg.Tests.Command = []string{"go", "test", "./..."}
	cfg.Tests.Timeout = 10_000_000_000
	cfg.Selection.Mode = "coverage"
	isolateArtifacts(&cfg, dir)
	cfg.Limits.MaxMutants = 1

	result, err := New(cfg).Run(context.Background(), RunRequest{Targets: []string{dir}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Mutants) != 1 {
		t.Fatalf("mutants = %d, want 1", len(result.Mutants))
	}
	if got := strings.Join(result.Mutants[0].TestCommand, " "); got != "go test ." {
		t.Fatalf("coverage selection command = %q, want package command", got)
	}
	if _, err := os.Stat(cfg.Selection.CoverageProfile); err != nil {
		t.Fatalf("coverage profile was not written: %v", err)
	}
	data, err := os.ReadFile(cfg.Selection.TimingsPath)
	if err != nil {
		t.Fatalf("timing history was not written: %v", err)
	}
	var timings map[string]int64
	if err := json.Unmarshal(data, &timings); err != nil {
		t.Fatalf("timing history is not JSON: %v", err)
	}
	if timings[result.Mutants[0].MutantID] <= 0 {
		t.Fatalf("timing not recorded for mutant: %#v", timings)
	}
}

func TestCacheKeyChangesWhenGoModOrTestsChange(t *testing.T) {
	dir := writeFixture(t)
	cfg := config.Defaults()
	cfg.Tests.Command = []string{"go", "test", "./..."}
	isolateArtifacts(&cfg, dir)

	e := New(cfg)
	discovered, err := e.discoverForTest([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	mutants, err := e.generateMutants(discovered)
	if err != nil {
		t.Fatal(err)
	}
	if len(mutants) == 0 {
		t.Fatal("no mutants generated")
	}
	first, err := e.cacheKeyForTest(mutants[0], TestPlan{Command: []string{"go", "test", "."}})
	if err != nil {
		t.Fatal(err)
	}
	testPath := filepath.Join(dir, "calc_test.go")
	f, err := os.OpenFile(testPath, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n// new assertion coming later\n"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := e.cacheKeyForTest(mutants[0], TestPlan{Command: []string{"go", "test", "."}})
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("cache key did not change after relevant test file changed")
	}
}
