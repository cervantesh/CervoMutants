package report

import (
	"fmt"

	"github.com/cervantesh/cervo-mutants/pkg/engine"
)

type laneInfo struct {
	label    string
	detail   string
	guidance string
}

func reportLaneInterpretation(result engine.RunResult) (laneInfo, bool) {
	summary := result.Summary
	actionable := summary.Actionable
	health := summary.DenominatorHealth

	switch {
	case actionable.TrueActionableSurvivors == 0 && len(health.Warnings) > 0:
		return laneInfo{
			label:    "retargeting signal",
			detail:   "the run completed, but denominator pressure dominates and this bounded slice did not produce immediate review work",
			guidance: "keep the artifact and retarget the next run to a hotter package, subtree, or shard before judging broader rollout fit",
		}, true
	case actionable.TrueActionableSurvivors == 0 && summary.Survived == 0 && (health.Healthy || len(health.Warnings) == 0):
		return laneInfo{
			label:    "healthy no-action lane",
			detail:   "this bounded slice produced understandable denominator health and no immediate survivor work",
			guidance: "keep the artifact; widen or retarget only if you need more review pressure from a different slice",
		}, true
	case actionable.CollapsedSemanticDuplicates > 0 || (actionable.SemanticGroupReviewUnits > 0 && summary.Survived > actionable.TrueActionableSurvivors):
		groupLabel := "groups"
		if actionable.SemanticGroupReviewUnits == 1 {
			groupLabel = "group"
		}
		return laneInfo{
			label:    "grouped review lane",
			detail:   fmt.Sprintf("%d raw survivors collapsed into %d immediate review units across %d semantic %s", summary.Survived, actionable.TrueActionableSurvivors, actionable.SemanticGroupReviewUnits, groupLabel),
			guidance: "review the grouped equivalent-risk family once before splitting it into multiple separate test tasks",
		}, true
	case actionable.TrueActionableSurvivors > 0:
		return laneInfo{
			label:    "direct review lane",
			detail:   "raw survivors and immediate review workload are closely aligned in this bounded slice",
			guidance: "start with the top survivor queue and nearby-test hints before widening the target or changing policy depth",
		}, true
	default:
		return laneInfo{}, false
	}
}
