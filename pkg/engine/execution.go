package engine

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/cervantesh/cervo-mutants/pkg/config"
	"github.com/cervantesh/cervo-mutants/pkg/isolate"
)

func (e *Engine) runMutants(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	if e.cfg.Execution.Resume {
		return e.newRunSession().runMutantsWithResume(ctx, mutants, quarantined)
	}
	return e.newRunSession().runMutants(ctx, mutants, quarantined)
}

func (s *runSession) runMutants(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	if s.engine.cfg.Execution.Resume {
		return s.runMutantsWithResume(ctx, mutants, quarantined)
	}
	workers := s.workerCount(len(mutants))
	if workers <= 1 {
		return s.runMutantsSerial(ctx, mutants, quarantined)
	}
	return s.runMutantsParallel(ctx, mutants, quarantined, workers)
}

func (s *runSession) runMutantsWithResume(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	completed, err := s.loadPartialResults(mutants)
	if err != nil {
		return nil, err
	}
	if len(completed) == 0 {
		workers := s.workerCount(len(mutants))
		if workers <= 1 {
			return s.runMutantsSerial(ctx, mutants, quarantined)
		}
		return s.runMutantsParallel(ctx, mutants, quarantined, workers)
	}
	results := make([]MutantResult, 0, len(mutants))
	remaining := make([]Mutant, 0, len(mutants))
	for _, mutant := range mutants {
		if result, ok := completed[mutant.ID]; ok {
			result = refreshCachedMutantResult(result, mutant)
			result.PreviousStatus = result.Status
			result.Status = StatusCached
			result.StatusReason = "result reused from partial checkpoint"
			results = append(results, result)
			continue
		}
		remaining = append(remaining, mutant)
	}
	next, err := s.runMutantsSerial(ctx, remaining, quarantined)
	if err != nil {
		return nil, err
	}
	results = append(results, next...)
	return orderResults(mutants, results), nil
}

func (s *runSession) runMutantsSerial(ctx context.Context, mutants []Mutant, quarantined map[string]bool) ([]MutantResult, error) {
	results := make([]MutantResult, 0, len(mutants))
	start := s.clockNow()
	for i, mutant := range mutants {
		if quarantined[mutant.ID] {
			result := MutantResult{MutantID: mutant.ID, Status: StatusQuarantined, StatusReason: "mutant is in active quarantine", Mutant: mutant}
			results = append(results, result)
			s.recordProgress(start, i+1, len(mutants), result)
			s.writePartialResults(results)
			continue
		}
		if result, ok := s.suppressedResult(mutant); ok {
			results = append(results, result)
			s.recordProgress(start, i+1, len(mutants), result)
			s.writePartialResults(results)
			continue
		}
		if s.budgetExhausted(start) {
			result := MutantResult{MutantID: mutant.ID, Status: StatusPendingBudget, FailureKind: "budget_exhausted", StatusReason: "budget exhausted before mutant execution", Mutant: mutant}
			results = append(results, result)
			s.recordProgress(start, i+1, len(mutants), result)
			s.writePartialResults(results)
			continue
		}
		mutantResult, err := s.runMutant(ctx, mutant)
		if err != nil {
			return nil, err
		}
		results = append(results, mutantResult)
		s.recordProgress(start, i+1, len(mutants), mutantResult)
		s.writePartialResults(results)
	}
	return results, nil
}

type indexedMutant struct {
	index  int
	mutant Mutant
}

type indexedResult struct {
	index  int
	result MutantResult
	err    error
}

func (s *runSession) runMutantsParallel(ctx context.Context, mutants []Mutant, quarantined map[string]bool, workers int) ([]MutantResult, error) {
	results := make([]MutantResult, len(mutants))
	jobs := make(chan indexedMutant, len(mutants))
	done := make(chan indexedResult, len(mutants))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	startParallelWorkers(ctx, workers, jobs, done, s.runMutant)
	start := s.clockNow()
	dispatchParallelJobs(s, mutants, quarantined, results, jobs, start)
	return s.collectParallelResults(done, results, len(mutants), start, cancel)
}

func startParallelWorkers(ctx context.Context, workers int, jobs <-chan indexedMutant, done chan<- indexedResult, run func(context.Context, Mutant) (MutantResult, error)) {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if ctx.Err() != nil {
					done <- indexedResult{index: job.index, err: ctx.Err()}
					continue
				}
				result, err := run(ctx, job.mutant)
				done <- indexedResult{index: job.index, result: result, err: err}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(done)
	}()
}

func dispatchParallelJobs(s *runSession, mutants []Mutant, quarantined map[string]bool, results []MutantResult, jobs chan<- indexedMutant, start time.Time) {
	for i, mutant := range mutants {
		if quarantined[mutant.ID] {
			results[i] = MutantResult{MutantID: mutant.ID, Status: StatusQuarantined, StatusReason: "mutant is in active quarantine", Mutant: mutant}
			continue
		}
		if result, ok := s.suppressedResult(mutant); ok {
			results[i] = result
			continue
		}
		if s.budgetExhausted(start) {
			results[i] = MutantResult{MutantID: mutant.ID, Status: StatusPendingBudget, FailureKind: "budget_exhausted", StatusReason: "budget exhausted before mutant execution", Mutant: mutant}
			continue
		}
		jobs <- indexedMutant{index: i, mutant: mutant}
	}
	close(jobs)
}

func (s *runSession) collectParallelResults(done <-chan indexedResult, results []MutantResult, total int, start time.Time, cancel context.CancelFunc) ([]MutantResult, error) {
	var firstErr error
	completed := 0
	for item := range done {
		if item.err != nil && firstErr == nil {
			firstErr = item.err
			cancel()
		}
		results[item.index] = item.result
		completed++
		s.recordProgress(start, completed, total, item.result)
		s.writePartialResults(compactedResults(results))
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return results, nil
}

func (s *runSession) budgetExhausted(start time.Time) bool {
	return s.engine.cfg.Execution.Budget > 0 && s.elapsedSince(start) >= s.engine.cfg.Execution.Budget
}

func refreshCachedMutantResult(result MutantResult, mutant Mutant) MutantResult {
	result.MutantID = mutant.ID
	result.Mutant = mutant
	result.SuggestedSkipReason = mutant.SuggestedSkipReason
	result.SuggestedTestScope = suggestedTestScope(mutant)
	result.NearestTests = append([]string(nil), mutant.NearbyTests...)
	return result
}

func (s *runSession) suppressedResult(mutant Mutant) (MutantResult, bool) {
	rule, ok := strongestSuppression(mutant.SuppressionAudit)
	if !ok || rule.Action != "suppress" {
		return MutantResult{}, false
	}
	return MutantResult{
		MutantID:           mutant.ID,
		Status:             StatusIgnored,
		StatusReason:       fmt.Sprintf("suppressed by audited rule %q: %s", rule.Name, rule.Reason),
		Mutant:             mutant,
		SuggestedTestScope: suggestedTestScope(mutant),
		NearestTests:       mutant.NearbyTests,
	}, true
}

func strongestSuppression(audits []SuppressionAudit) (SuppressionAudit, bool) {
	var best SuppressionAudit
	bestPriority := -1
	for _, audit := range audits {
		priority := suppressionPriority(audit.Action)
		if priority > bestPriority {
			best = audit
			bestPriority = priority
		}
	}
	return best, bestPriority >= 0
}

func suppressionPriority(action string) int {
	switch action {
	case config.SuppressionReportOnly:
		return 0
	case config.SuppressionLowerPriority:
		return 1
	case "quarantine-required":
		return 2
	case "suppress":
		return 3
	default:
		return -1
	}
}

func (e *Engine) workerCount(mutants int) int {
	return effectiveWorkerCount(runtime.GOOS, e.cfg.Execution.Isolation, e.cfg.Execution.Workers, mutants)
}

func (e *Engine) environment(mutants int) Environment {
	wd, _ := os.Getwd()
	tempPlan := isolate.ResolveTempRoot(wd, e.cfg.Execution.TempRoot)
	runtimePlan := effectiveTestCommandEnv(runtime.GOOS, e.cfg.Execution.Isolation, e.workerCount(mutants), e.cfg.Tests.Command, os.Environ())
	env := Environment{
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		GoVersion:       runtime.Version(),
		WorkingDir:      wd,
		TempDir:         os.TempDir(),
		TempRoot:        tempPlan.Root,
		Isolation:       e.cfg.Execution.Isolation,
		Workers:         e.workerCount(mutants),
		TestTimeout:     e.cfg.Tests.Timeout.String(),
		GoFlags:         os.Getenv("GOFLAGS"),
		GoMaxProcs:      os.Getenv("GOMAXPROCS"),
		GoMemLimit:      os.Getenv("GOMEMLIMIT"),
		CI:              os.Getenv("CI"),
		WSL:             isWSL(),
		CGroup:          cgroupSummary(),
		WindowsOneDrive: runtime.GOOS == "windows" && pathMentionsOneDrive(wd),
		Warnings:        append([]string{}, tempPlan.Warnings...),
	}
	if e.cfg.Execution.Resources.MaxProcessMemoryMB > 0 || e.cfg.Execution.Resources.MaxProcesses > 0 || tempPlan.Source != "" || runtimePlan.Applied {
		env.Extra = map[string]string{}
		if e.cfg.Execution.Resources.MaxProcessMemoryMB > 0 {
			env.Extra["max_process_memory_mb"] = strconv.Itoa(e.cfg.Execution.Resources.MaxProcessMemoryMB)
		}
		if e.cfg.Execution.Resources.MaxProcesses > 0 {
			env.Extra["max_processes"] = strconv.Itoa(e.cfg.Execution.Resources.MaxProcesses)
		}
		if tempPlan.Source != "" {
			env.Extra["temp_root_source"] = tempPlan.Source
		}
		if runtimePlan.Applied {
			env.Extra["effective_goflags"] = runtimePlan.GoFlags
			env.Extra["effective_gomaxprocs"] = runtimePlan.GOMAXPROCS
		}
	}
	if e.cfg.Execution.Budget > 0 {
		env.Budget = e.cfg.Execution.Budget.String()
	}
	if runtime.GOOS == "windows" && e.cfg.Execution.Isolation == config.IsolationTempWorkdir && e.cfg.Execution.Workers > env.Workers {
		env.Warnings = append(env.Warnings, fmt.Sprintf("Windows temp-workdir worker cap applied: requested=%d effective=%d", e.cfg.Execution.Workers, env.Workers))
	}
	if runtime.GOOS == "windows" && e.cfg.Execution.Isolation == config.IsolationTempWorkdir && e.cfg.Tests.Timeout > 0 && e.cfg.Tests.Timeout < 20*time.Second {
		env.Warnings = append(env.Warnings, fmt.Sprintf("per-mutant timeout %s may be too aggressive for Windows temp-workdir runs", e.cfg.Tests.Timeout))
	}
	if runtime.GOOS != "windows" && hasProcessLimits(e.cfg.Execution.Resources) {
		env.Warnings = append(env.Warnings, "process resource limits are not enforced on this platform; continuing without process-limit isolation")
	}
	return env
}

func effectiveWorkerCount(goos, isolation string, requested, mutants int) int {
	workers := requested
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > mutants && mutants > 0 {
		workers = mutants
	}
	if workers < 1 {
		workers = 1
	}
	if goos == "windows" && isolation == config.IsolationTempWorkdir && workers > 2 {
		workers = 2
	}
	return workers
}
