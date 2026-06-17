package triage

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type TestRecommendation struct {
	Priority            string
	Strategy            string
	Summary             string
	CandidateTests      []string
	SuggestedAssertions []string
	Rationale           []string
}

func (r TestRecommendation) Empty() bool {
	return strings.TrimSpace(r.Priority) == "" &&
		strings.TrimSpace(r.Strategy) == "" &&
		strings.TrimSpace(r.Summary) == "" &&
		len(r.CandidateTests) == 0 &&
		len(r.SuggestedAssertions) == 0 &&
		len(r.Rationale) == 0
}

func BuildTestRecommendation(goos string, result Result, groupSize int, priority string) *TestRecommendation {
	if result.Status != StatusSurvived {
		return nil
	}
	orderedTests := orderedNearbyTests(result.Mutant)
	strategy := recommendationStrategy(goos, result)
	rec := TestRecommendation{
		Priority:            strings.TrimSpace(priority),
		Strategy:            strategy,
		Summary:             recommendationSummary(result, orderedTests),
		CandidateTests:      orderedTests,
		SuggestedAssertions: recommendationAssertions(goos, result),
		Rationale:           recommendationRationale(goos, result, orderedTests, groupSize),
	}
	if rec.Empty() {
		return nil
	}
	return &rec
}

func orderedNearbyTests(mutant Mutant) []string {
	tests := append([]string{}, mutant.NearbyTests...)
	sort.SliceStable(tests, func(i, j int) bool {
		left := nearbyTestScore(mutant, tests[i])
		right := nearbyTestScore(mutant, tests[j])
		if left != right {
			return left > right
		}
		return tests[i] < tests[j]
	})
	return tests
}

func nearbyTestScore(mutant Mutant, path string) int {
	score := 0
	pathLower := strings.ToLower(filepath.ToSlash(path))
	fileBase := strings.ToLower(strings.TrimSuffix(filepath.Base(mutant.File), filepath.Ext(mutant.File)))
	if fileBase != "" && strings.Contains(pathLower, fileBase) {
		score += 6
	}
	functionKey := normalizedAlphaNum(mutant.Function)
	if functionKey != "" && strings.Contains(normalizedAlphaNum(pathLower), functionKey) {
		score += 8
	}
	if strings.Contains(strings.ToLower(mutant.GroupLabel), "sort") && strings.Contains(pathLower, "sort") {
		score += 4
	}
	if strings.Contains(strings.ToLower(mutant.Operator), "loop") && strings.Contains(pathLower, "loop") {
		score += 3
	}
	if strings.Contains(pathLower, "_test.go") {
		score++
	}
	return score
}

func recommendationStrategy(goos string, result Result) string {
	if strings.EqualFold(goos, "windows") && result.Mutant.PlatformSensitive {
		return "review-platform-contract"
	}
	switch result.Mutant.Operator {
	case "conditionals-boundary", "conditionals-negation", "logical", "boolean-literals", "slice-map-len-boundary":
		return "tighten-branch-assertions"
	case "numeric-literals", "arithmetic-basic", "assignment-arithmetic", "return-bool-literals", "returns", "literals", "string-empty-literals":
		return "tighten-value-assertions"
	case "nil-checks", "error-returns":
		return "assert-error-contracts"
	case "inc-dec", "loop-control":
		return "assert-termination-and-progress"
	default:
		return "add-targeted-regression-test"
	}
}

func recommendationSummary(result Result, orderedTests []string) string {
	action := suggestionHint(result)
	target := SuggestedTestScope(result.Mutant)
	if len(orderedTests) > 0 {
		switch result.HistoryStatus {
		case "long_standing_survivor":
			return fmt.Sprintf("Promote `%s` into a named regression: %s", orderedTests[0], action)
		case "new_survivor":
			return fmt.Sprintf("Start with `%s` while this survivor is new: %s", orderedTests[0], action)
		default:
			return fmt.Sprintf("Start with `%s`: %s", orderedTests[0], action)
		}
	}
	if result.HistoryStatus == "long_standing_survivor" {
		return fmt.Sprintf("Add a named regression test in `%s`: %s", target, action)
	}
	return fmt.Sprintf("Add a focused test in `%s`: %s", target, action)
}

func suggestionHint(result Result) string {
	if hint := cleanSentence(result.Mutant.Hint); hint != "" {
		return hint
	}
	switch result.Mutant.Operator {
	case "conditionals-boundary", "conditionals-negation", "logical", "boolean-literals", "slice-map-len-boundary":
		return "Assert the exact boundary or branch outcome, not only a broad success path."
	case "numeric-literals", "arithmetic-basic", "assignment-arithmetic", "return-bool-literals", "returns", "literals", "string-empty-literals":
		return "Assert the exact returned value or emitted output, not only presence or non-error behavior."
	case "nil-checks", "error-returns":
		return "Assert the concrete error and nil contract at the call boundary."
	case "inc-dec", "loop-control":
		return "Assert termination, progress, and final collection size directly."
	default:
		if description := cleanSentence(result.Mutant.Description); description != "" {
			return description
		}
		return "Add a focused assertion around the mutated behavior."
	}
}

func recommendationAssertions(goos string, result Result) []string {
	assertions := make([]string, 0, 3)
	assertions = appendUniqueSuggestion(assertions, suggestionHint(result))
	if coverageSuggestion := coverageAssertion(result.CoverageSource); coverageSuggestion != "" {
		assertions = appendUniqueSuggestion(assertions, coverageSuggestion)
	}
	if strings.EqualFold(goos, "windows") && result.Mutant.PlatformSensitive {
		assertions = appendUniqueSuggestion(assertions, "Exercise the permission-mode behavior on Windows explicitly before treating the survivor as actionable.")
	}
	if result.HistoryStatus == "long_standing_survivor" {
		assertions = appendUniqueSuggestion(assertions, "Give the fix a dedicated regression case name so future runs expose the contract immediately.")
	}
	return assertions
}

func recommendationRationale(goos string, result Result, orderedTests []string, groupSize int) []string {
	rationale := make([]string, 0, 6)
	rationale = append(rationale, fmt.Sprintf("operator=%s -> %s", result.Mutant.Operator, operatorFamilyReason(result.Mutant.Operator)))
	rationale = append(rationale, fmt.Sprintf("coverage_source=%s -> %s", result.CoverageSource, coverageReason(result.CoverageSource)))
	if len(orderedTests) > 0 {
		rationale = append(rationale, fmt.Sprintf("nearby_tests=%d -> start with %s", len(orderedTests), orderedTests[0]))
	} else {
		rationale = append(rationale, fmt.Sprintf("nearby_tests=0 -> create a focused test in %s", SuggestedTestScope(result.Mutant)))
	}
	if result.HistoryStatus != "" {
		rationale = append(rationale, fmt.Sprintf("history=%s -> %s", result.HistoryStatus, historyReason(result.HistoryStatus, result.SurvivorAgeRuns)))
	}
	if groupSize > 1 && result.Mutant.GroupLabel != "" {
		rationale = append(rationale, fmt.Sprintf("semantic_group=%s -> one good review can collapse %d similar survivors", result.Mutant.GroupLabel, groupSize))
	}
	if strings.EqualFold(goos, "windows") && result.Mutant.PlatformSensitive {
		rationale = append(rationale, "goos=windows -> permission-mode mutations need target-OS verification before escalation")
	}
	return rationale
}

func coverageAssertion(source string) string {
	switch {
	case strings.Contains(source, "fallback"):
		return "Keep the assertion file-local; fallback coverage usually means the package-level run is too broad."
	case strings.Contains(source, "package-mode"):
		return "Prefer a focused assertion that proves the mutated branch, not only package-level execution."
	case source == "coverage-mode":
		return ""
	default:
		return ""
	}
}

func coverageReason(source string) string {
	switch {
	case source == "":
		return "coverage signal unavailable; rely on the closest nearby test first"
	case strings.Contains(source, "fallback"):
		return "the mutant was reached through fallback coverage rather than a tight file-level match"
	case strings.Contains(source, "package-mode"):
		return "the current signal came from package-level execution and may still miss a precise assertion"
	case strings.Contains(source, "coverage-mode"):
		return "the mutant was matched by coverage data, so the next test should usually be an assertion upgrade"
	default:
		return "use the current selection signal as a starting point, then tighten the assertion"
	}
}

func operatorFamilyReason(operator string) string {
	switch operator {
	case "conditionals-boundary", "conditionals-negation", "logical", "boolean-literals", "slice-map-len-boundary":
		return "branch and boundary assertions usually kill this operator family"
	case "numeric-literals", "arithmetic-basic", "assignment-arithmetic", "return-bool-literals", "returns", "literals", "string-empty-literals":
		return "exact value checks usually kill this operator family"
	case "nil-checks", "error-returns":
		return "error and nil contracts usually kill this operator family"
	case "inc-dec", "loop-control":
		return "termination and progress checks usually kill this operator family"
	default:
		return "use a focused regression assertion around the mutated behavior"
	}
}

func historyReason(status string, runs int) string {
	switch status {
	case "new_survivor":
		return "fix the closest nearby test while the regression is still fresh"
	case "long_standing_survivor":
		return fmt.Sprintf("this mutant has survived %d runs; convert the next fix into a named regression", runs)
	case "existing_survivor":
		return "tighten the closest existing test before broadening coverage"
	default:
		return "use history as a prioritization hint, not as proof of equivalence"
	}
}

func appendUniqueSuggestion(values []string, value string) []string {
	value = cleanSentence(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return values
		}
	}
	return append(values, value)
}

func cleanSentence(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimRight(value, ".")
	if value == "" {
		return ""
	}
	return value + "."
}

func normalizedAlphaNum(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
