package mutator

import (
	"strings"
	"testing"
)

func TestConservativeFastMutatorsStayHighSignal(t *testing.T) {
	src := `package sample

func Check(n int, ready bool, p *int) bool {
	if n < 1 && ready && p == nil {
		return n + 1 > 0
	}
	return false
}
`

	mutants, err := Generate("sample", "sample.go", []byte(src), ProfileConservativeFast)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	operators := operatorSet(mutants)
	for _, want := range []string{"conditionals-negation", "conditionals-boundary", "arithmetic-basic"} {
		if !operators[want] {
			t.Fatalf("conservative-fast missing %s: %+v", want, mutants)
		}
	}
	for _, noisy := range []string{"logical", "boolean-literals", "nil-checks"} {
		if operators[noisy] {
			t.Fatalf("conservative-fast generated noisy operator %s: %+v", noisy, mutants)
		}
	}
}

func TestConservativeMutatorsGenerateStableActionableMutants(t *testing.T) {
	src := `package sample

func Check(n int, ready bool, p *int) bool {
	if n == 1 && ready && p == nil {
		return true
	}
	return false
}
`

	mutants, err := Generate("sample", "sample.go", []byte(src), ProfileConservative)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if len(mutants) < 4 {
		t.Fatalf("generated %d mutants, want at least 4", len(mutants))
	}

	seen := map[string]bool{}
	for _, mutant := range mutants {
		if mutant.ID == "" || mutant.Operator == "" || mutant.Diff == "" || mutant.Line == 0 {
			t.Fatalf("mutant missing actionable fields: %+v", mutant)
		}
		if mutant.Description == "" {
			t.Fatalf("mutant missing description: %+v", mutant)
		}
		if seen[mutant.ID] {
			t.Fatalf("duplicate mutant ID: %s", mutant.ID)
		}
		seen[mutant.ID] = true
	}

	foundConditional := false
	foundLogical := false
	foundBoolean := false
	for _, mutant := range mutants {
		foundConditional = foundConditional || mutant.Operator == "conditionals-negation"
		foundLogical = foundLogical || mutant.Operator == "logical"
		foundBoolean = foundBoolean || mutant.Operator == "boolean-literals"
		if strings.Contains(mutant.Diff, "--- sample.go") && strings.Contains(mutant.Diff, "+++ sample.go") {
			continue
		}
		t.Fatalf("mutant diff is not unified enough: %q", mutant.Diff)
	}
	if !foundConditional || !foundLogical || !foundBoolean {
		t.Fatalf("missing expected operators: conditionals=%v logical=%v boolean=%v", foundConditional, foundLogical, foundBoolean)
	}
	if operatorSet(mutants)["nil-checks"] {
		t.Fatalf("conservative should not generate nil-checks after Cobra noise study: %+v", mutants)
	}
}

func TestDefaultProfileAddsNilChecks(t *testing.T) {
	src := `package sample

func Check(p *int) bool {
	return p == nil
}
`
	mutants, err := Generate("sample", "sample.go", []byte(src), ProfileDefault)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !operatorSet(mutants)["nil-checks"] {
		t.Fatalf("default profile missing nil-checks: %+v", mutants)
	}
}

func TestAggressiveProfileAddsLiteralAndReturnMutators(t *testing.T) {
	src := `package sample

func Answer() int {
	return 1
}

func Ready() bool {
	return true
}
`

	conservative, err := Generate("sample", "sample.go", []byte(src), ProfileConservative)
	if err != nil {
		t.Fatalf("Generate conservative returned error: %v", err)
	}
	for _, mutant := range conservative {
		if mutant.Operator == "literals" || mutant.Operator == "returns" {
			t.Fatalf("conservative generated aggressive mutant: %+v", mutant)
		}
	}

	aggressive, err := Generate("sample", "sample.go", []byte(src), ProfileAggressive)
	if err != nil {
		t.Fatalf("Generate aggressive returned error: %v", err)
	}
	foundLiteral := false
	foundReturn := false
	for _, mutant := range aggressive {
		foundLiteral = foundLiteral || mutant.Operator == "literals"
		foundReturn = foundReturn || mutant.Operator == "returns"
	}
	if !foundLiteral || !foundReturn {
		t.Fatalf("aggressive profile missing literal/return mutants: literal=%v return=%v mutants=%+v", foundLiteral, foundReturn, aggressive)
	}
}

func TestInlineIgnoreRequiresReasonWhenConfigured(t *testing.T) {
	src := `package sample

func Check(n int) bool {
	// cervomut:ignore conditionals reason="covered by generated contract"
	return n == 1
}
`
	mutants, err := Generate("sample", "sample.go", []byte(src), ProfileConservative)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	for _, mutant := range mutants {
		if strings.HasPrefix(mutant.Operator, "conditionals-") {
			t.Fatalf("conditionals mutant was not ignored: %+v", mutant)
		}
	}

	bad := []byte(strings.Replace(src, ` reason="covered by generated contract"`, "", 1))
	if _, err := ValidateInlineIgnores("sample.go", bad, true); err == nil {
		t.Fatal("ValidateInlineIgnores accepted ignore without reason")
	}
}

func operatorSet(mutants []Mutant) map[string]bool {
	operators := map[string]bool{}
	for _, mutant := range mutants {
		operators[mutant.Operator] = true
	}
	return operators
}
