package report

import (
	"fmt"
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/engine"
)

func TestRecommendations(result engine.RunResult) string {
	var b strings.Builder
	survivors := rankedSurvivors(result.Mutants)
	visible, stats := filterSurvivors(result, survivors, SurvivorsOptions{ActionableOnly: true})
	fmt.Fprintf(&b, "# CervoMutants Test Recommendations\n\n")
	fmt.Fprintf(&b, "Actionable review units: **%d** of **%d** survivors (filtered=%d collapsed=%d)\n\n", stats.Shown, stats.Total, stats.Filtered, stats.CollapsedGroup)
	if len(visible) == 0 {
		b.WriteString("No actionable survivors currently need a test recommendation.\n")
		return b.String()
	}
	for _, mutant := range visible {
		fmt.Fprintf(&b, "## #%d `%s`\n\n", mutant.SurvivorRank, mutant.MutantID)
		fmt.Fprintf(&b, "- Location: `%s:%d`\n", mutant.Mutant.File, mutant.Mutant.Line)
		fmt.Fprintf(&b, "- Operator: `%s`\n", mutant.Mutant.Operator)
		fmt.Fprintf(&b, "- Actionability: `%s`\n", fallbackText(mutant.Actionability, "unknown"))
		if ownership := ownershipRouteSummary(mutant.Mutant.Ownership); ownership != "" {
			fmt.Fprintf(&b, "- Ownership: `%s`\n", ownership)
		}
		fmt.Fprintf(&b, "- Suggested scope: `%s`\n", fallbackText(mutant.SuggestedTestScope, "."))
		if recommendation := mutant.TestRecommendation; recommendation != nil {
			fmt.Fprintf(&b, "- Recommendation priority: `%s`\n", fallbackText(recommendation.Priority, "unknown"))
			fmt.Fprintf(&b, "- Recommendation strategy: `%s`\n", fallbackText(recommendation.Strategy, "unknown"))
			if summary := strings.TrimSpace(recommendation.Summary); summary != "" {
				fmt.Fprintf(&b, "- Summary: %s\n", summary)
			}
			if len(recommendation.CandidateTests) > 0 {
				fmt.Fprintf(&b, "- Candidate tests: `%s`\n", strings.Join(recommendation.CandidateTests, "`, `"))
			}
			for _, assertion := range recommendation.SuggestedAssertions {
				fmt.Fprintf(&b, "- Suggested assertion: %s\n", assertion)
			}
			for _, rationale := range recommendation.Rationale {
				fmt.Fprintf(&b, "- Rationale: %s\n", rationale)
			}
		}
		if skip := strings.TrimSpace(mutant.SuggestedSkipReason); skip != "" {
			fmt.Fprintf(&b, "- Skip guidance: %s\n", skip)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func recommendationPrimaryTest(recommendation *engine.TestRecommendation) string {
	if recommendation == nil || len(recommendation.CandidateTests) == 0 {
		return ""
	}
	return recommendation.CandidateTests[0]
}

func recommendationSummary(recommendation *engine.TestRecommendation) string {
	if recommendation == nil {
		return ""
	}
	return strings.TrimSpace(recommendation.Summary)
}

func recommendationStrategy(recommendation *engine.TestRecommendation) string {
	if recommendation == nil {
		return ""
	}
	return strings.TrimSpace(recommendation.Strategy)
}

func recommendationAssertions(recommendation *engine.TestRecommendation) string {
	if recommendation == nil || len(recommendation.SuggestedAssertions) == 0 {
		return ""
	}
	return strings.Join(recommendation.SuggestedAssertions, " | ")
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
