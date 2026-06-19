package report

import "github.com/cervantesh/cervo-mutants/pkg/engine"

func denominatorGuidanceLines(health engine.DenominatorHealth) []string {
	if !hasLowSignalDenominatorWarning(health.Warnings) {
		return nil
	}
	return []string{
		"Preserve this report and treat the run as target-selection feedback before changing score expectations.",
		"Retarget the next run to a hotter package, subtree, or bounded shard before widening to ./....",
		"Rerun on the narrower target before judging recommendation quality or broader rollout fit.",
	}
}

func hasLowSignalDenominatorWarning(warnings []string) bool {
	for _, warning := range warnings {
		switch warning {
		case "no_effective_mutants", "not_covered_exceeds_effective", "score_denominator_dwarfs_effective", "high_score_poor_denominator_health":
			return true
		}
	}
	return false
}
