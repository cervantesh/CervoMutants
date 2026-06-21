package engine

import (
	"context"
	"sync"
	"time"
)

type engineDeps struct {
	discoverMutants func(*Engine, []string) ([]Mutant, error)
	runBaseline     func(*runSession, context.Context, []string) (MutantResult, error)
	runMutants      func(*runSession, context.Context, []Mutant, map[string]bool) ([]MutantResult, error)
	writeReports    func(*Engine, RunResult) error
}

func defaultEngineDeps() engineDeps {
	return engineDeps{
		discoverMutants: func(e *Engine, targets []string) ([]Mutant, error) {
			return e.discoverMutants(targets)
		},
		runBaseline: func(s *runSession, ctx context.Context, targets []string) (MutantResult, error) {
			return s.runBaseline(ctx, targets)
		},
		runMutants: func(s *runSession, ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
			return s.runMutants(ctx, mutants, quarantined)
		},
		writeReports: func(e *Engine, result RunResult) error {
			return e.writeReports(result)
		},
	}
}

type runSession struct {
	engine          *Engine
	timingMu        sync.Mutex
	checkpointMu    sync.Mutex
	checkpointScope []Mutant
	sliceMeta       SliceMetadata
	coverageBaseDir string
}

func (e *Engine) newRunSession() *runSession {
	return &runSession{engine: e}
}

func (s *runSession) clockNow() time.Time {
	return s.engine.clockNow()
}

func (s *runSession) elapsedSince(start time.Time) time.Duration {
	return s.engine.elapsedSince(start)
}

func (s *runSession) workerCount(mutants int) int {
	return s.engine.workerCount(mutants)
}

func (s *runSession) environment(mutants int) Environment {
	return s.engine.environment(mutants)
}

func (s *runSession) applySurvivorRanking(results []MutantResult) {
	s.engine.applySurvivorRanking(results)
}
