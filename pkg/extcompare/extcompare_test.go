package extcompare

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCervoReportNormalizesMetrics(t *testing.T) {
	path := writeJSON(t, `{
  "summary": {"total": 3, "killed": 1, "survived": 1, "not_covered": 1, "score": 50}
}`)

	result, err := ParseCervo(path)
	if err != nil {
		t.Fatalf("ParseCervo returned error: %v", err)
	}
	if result.Tool != "cervo-mutant" || result.Total != 3 || result.Killed != 1 || result.Survived != 1 || result.NotCovered != 1 || result.Score != 50 {
		t.Fatalf("unexpected normalized result: %+v", result)
	}
}

func TestParseGremlinsReportNormalizesMetrics(t *testing.T) {
	path := writeJSON(t, `{
  "mutants_total": 4,
  "mutants_killed": 2,
  "mutants_lived": 1,
  "mutants_not_covered": 1,
  "mutants_not_viable": 0,
  "test_efficacy": 66.6667
}`)

	result, err := ParseGremlins(path)
	if err != nil {
		t.Fatalf("ParseGremlins returned error: %v", err)
	}
	if result.Tool != "gremlins" || result.Total != 4 || result.Killed != 2 || result.Survived != 1 || result.NotCovered != 1 {
		t.Fatalf("unexpected normalized result: %+v", result)
	}
	if result.Status != "ok" || result.TestEfficacy == 0 || result.DenominatorHealth.Effective != 3 {
		t.Fatalf("expected status, efficacy, and denominator health: %+v", result)
	}
}

func TestParseGremlinsReportClassifiesAllTimedOutAndPoorDenominatorHealth(t *testing.T) {
	path := writeJSON(t, `{
  "mutants_total": 0,
  "mutants_killed": 0,
  "mutants_lived": 0,
  "mutants_not_covered": 5,
  "test_efficacy": 0,
  "files": [
    {"mutations": [{"status": "TIMED OUT"}, {"status": "TIMED OUT"}]}
  ]
}`)

	result, err := ParseGremlins(path)
	if err != nil {
		t.Fatalf("ParseGremlins returned error: %v", err)
	}
	if result.Status != "all_timed_out" || result.TimedOut != 2 {
		t.Fatalf("unexpected Gremlins status: %+v", result)
	}
}

func TestNormalizeGremlinsPackageRootTargetMarksNotComparable(t *testing.T) {
	effective, notComparable := NormalizeGremlinsTarget("./...", "gremlins-package-root")
	if effective != "." || !notComparable {
		t.Fatalf("effective=%q notComparable=%t, want . true", effective, notComparable)
	}
	result := ApplyTarget(ToolResult{Tool: "gremlins", Status: "ok"}, "./...", effective, notComparable)
	if result.Target != "./..." || result.EffectiveTarget != "." || !result.NotComparable || len(result.Notes) == 0 {
		t.Fatalf("target metadata not applied: %+v", result)
	}
}

func TestDenominatorHealthWarnsWhenScoreHidesTimeouts(t *testing.T) {
	result := ToolResult{Tool: "gremlins", Status: "ok", Total: 3, Killed: 3, TimedOut: 1244, NotCovered: 37, Score: 100}
	result.DenominatorHealth = denominatorHealth(result)
	warnings := strings.Join(result.DenominatorHealth.Warnings, ",")
	for _, want := range []string{"timed_out_exceeds_effective", "not_covered_exceeds_effective", "high_score_poor_denominator_health"} {
		if !strings.Contains(warnings, want) {
			t.Fatalf("warnings missing %q: %+v", want, result.DenominatorHealth)
		}
	}
}

func TestParseGomuReportAcceptsJSONStatusResults(t *testing.T) {
	path := writeJSON(t, `{
  "totalMutants": 5,
  "results": [
    {"status": "KILLED"},
    {"status": "KILLED"},
    {"status": "SURVIVED"},
    {"status": "ERROR"},
    {"status": "NOT_VIABLE"}
  ]
}`)

	result, err := ParseGomu(path)
	if err != nil {
		t.Fatalf("ParseGomu returned error: %v", err)
	}
	if result.Tool != "gomu" || result.Total != 5 || result.Killed != 2 || result.Survived != 1 || result.Errors != 1 || result.NotViable != 1 {
		t.Fatalf("unexpected normalized result: %+v", result)
	}
}

func TestParseGoMutestingReportAcceptsJSONStats(t *testing.T) {
	path := writeJSON(t, `{
  "stats": {
    "totalMutantsCount": 4,
    "killedCount": 3,
    "escapedCount": 1,
    "notCoveredCount": 0,
    "errorCount": 0,
    "skippedCount": 0,
    "timeOutCount": 0,
    "msi": 0.75
  }
}`)

	result, err := ParseGoMutesting(path)
	if err != nil {
		t.Fatalf("ParseGoMutesting returned error: %v", err)
	}
	if result.Tool != "go-mutesting" || result.Total != 4 || result.Killed != 3 || result.Survived != 1 || result.Score != 75 {
		t.Fatalf("unexpected normalized result: %+v", result)
	}
}

func TestWriteStudy(t *testing.T) {
	out := filepath.Join(t.TempDir(), "study.json")
	if err := Write(out, []ToolResult{{Tool: "cervo-mutant", Completed: true, Total: 1}}); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("study missing: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("study file is empty")
	}
}

func writeJSON(t *testing.T, text string) string {
	t.Helper()
	return writeText(t, text)
}

func writeText(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
