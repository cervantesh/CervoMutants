package engine

import (
	"context"
	"strings"
	"time"

	internalgotestenv "github.com/cervantesh/cervo-mutants/pkg/internal/gotestenv"
)

func (e *Engine) selectTests(mutant Mutant) TestPlan {
	return e.newRunSession().selectTests(mutant)
}

func (e *Engine) runTest(ctx context.Context, job MutantJob) (MutantResult, error) {
	return e.newRunSession().runTest(ctx, job)
}

func (e *Engine) prepareMutation(mutant Mutant, command []string) (string, []string, func(), error) {
	return e.newRunSession().prepareMutation(mutant, command)
}

func (e *Engine) runMutant(ctx context.Context, mutant Mutant) (MutantResult, error) {
	return e.newRunSession().runMutant(ctx, mutant)
}

func (e *Engine) runMutantsSerial(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	return e.newRunSession().runMutantsSerial(ctx, mutants, quarantined)
}

func (e *Engine) runMutantsWithResume(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	return e.newRunSession().runMutantsWithResume(ctx, mutants, quarantined)
}

func (e *Engine) runMutantsParallel(ctx context.Context, mutants []Mutant, quarantined map[string]bool, workers int) ([]MutantResult, error) {
	return e.newRunSession().runMutantsParallel(ctx, mutants, quarantined, workers)
}

func (e *Engine) collectParallelResults(done <-chan indexedResult, results []MutantResult, total int, start time.Time, cancel context.CancelFunc) ([]MutantResult, error) {
	return e.newRunSession().collectParallelResults(done, results, total, start, cancel)
}

func (e *Engine) getCached(key string) (MutantResult, bool, error) {
	return e.newRunSession().getCached(key)
}

func (e *Engine) putCached(key string, result MutantResult) error {
	return e.newRunSession().putCached(key, result)
}

func (e *Engine) loadBaseline() (RunResult, bool, error) {
	return e.newRunSession().loadBaseline()
}

func (e *Engine) loadQuarantine() (map[string]bool, int, error) {
	return e.newRunSession().loadQuarantine()
}

func (e *Engine) cacheKey(mutant Mutant, plan TestPlan) (string, error) {
	return e.newRunSession().cacheKey(mutant, plan)
}

func (e *Engine) recordTiming(mutantID string, duration time.Duration) {
	e.newRunSession().recordTiming(mutantID, duration)
}

func (e *Engine) writePartialResults(results []MutantResult) {
	e.newRunSession().writePartialResults(results)
}

func (e *Engine) loadPartialResults(mutants []Mutant) (map[string]MutantResult, error) {
	return e.newRunSession().loadPartialResults(mutants)
}

func isGoTestCommand(command []string) bool {
	return internalgotestenv.IsGoTestCommand(command)
}

func packageScopedCommand(command []string, pkg string) []string {
	return internalgotestenv.PackageScopedCommand(command, pkg)
}

func withCoverProfile(command []string, profile string) []string {
	return internalgotestenv.WithCoverProfile(command, profile)
}

func withOverlayFlag(command []string, overlayPath string) []string {
	return internalgotestenv.WithOverlayFlag(command, overlayPath)
}

func isGoTestFlagWithSeparateValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	switch arg {
	case "-run", "-bench", "-count", "-timeout", "-coverprofile", "-covermode", "-coverpkg", "-tags", "-cpu", "-parallel", "-shuffle":
		return true
	default:
		return false
	}
}
