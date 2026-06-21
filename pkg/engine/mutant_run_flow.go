package engine

import "context"

func (s *runSession) runMutant(ctx context.Context, mutant Mutant) (MutantResult, error) {
	plan := s.selectTests(mutant)
	if !plan.CoversMutant {
		return MutantResult{
			MutantID:           mutant.ID,
			Status:             StatusNotCovered,
			TestCommand:        plan.Command,
			StatusReason:       "coverage profile did not execute mutant file",
			SelectionReason:    plan.Reason,
			CoverageSource:     plan.CoverageSource,
			Mutant:             mutant,
			SuggestedTestScope: suggestedTestScope(mutant),
			NearestTests:       mutant.NearbyTests,
		}, nil
	}
	key, err := s.cacheKey(mutant, plan)
	if err != nil {
		return MutantResult{}, err
	}
	if s.engine.cfg.Cache.Enabled && s.engine.cfg.Cache.Mode != "off" {
		if cached, ok, err := s.getCached(key); err == nil && ok {
			result := refreshCachedMutantResult(cached, mutant)
			result.PreviousStatus = result.Status
			result.Status = StatusCached
			result.StatusReason = "result reused from incremental cache"
			return result, nil
		}
	}
	workdir, command, cleanup, err := s.prepareMutation(mutant, plan.Command)
	if err != nil {
		return MutantResult{}, err
	}
	defer cleanup()
	result, err := s.runTest(ctx, MutantJob{ID: mutant.ID, Mutant: mutant, WorkDir: workdir, TestCommand: command, Timeout: s.engine.cfg.Tests.Timeout.String()})
	result.SelectionReason = plan.Reason
	result.CoverageSource = plan.CoverageSource
	result.SuggestedTestScope = suggestedTestScope(mutant)
	result.NearestTests = mutant.NearbyTests
	applySemanticResultMetadata(&result)
	s.recordTiming(mutant.ID, result.Duration)
	if err == nil && s.engine.cfg.Cache.Enabled && s.engine.cfg.Cache.Mode == "incremental" {
		_ = s.putCached(key, result)
	}
	return result, err
}
