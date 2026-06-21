package engine

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	internalgotestenv "github.com/cervantesh/cervo-mutants/pkg/internal/gotestenv"
)

type testCommandEnvPlan = internalgotestenv.Plan

func effectiveTestCommandEnv(goos, isolation string, workers int, command, baseEnv []string) testCommandEnvPlan {
	return internalgotestenv.EffectiveCommandEnv(goos, isolation, workers, command, baseEnv)
}

func normalizeGoFlags(current string) string {
	return internalgotestenv.NormalizeGoFlags(current)
}

func (s *runSession) runTest(ctx context.Context, job MutantJob) (MutantResult, error) {
	timeout := s.engine.cfg.Tests.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if len(job.TestCommand) == 0 {
		return MutantResult{}, errors.New("test command is empty")
	}
	start := time.Now()
	cmd := exec.CommandContext(runCtx, job.TestCommand[0], job.TestCommand[1:]...)
	cmd.Dir = job.WorkDir
	envPlan := effectiveTestCommandEnv(runtime.GOOS, s.engine.cfg.Execution.Isolation, s.workerCount(0), job.TestCommand, os.Environ())
	if envPlan.Applied {
		cmd.Env = envPlan.Env
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Start()
	handle := noopProcessLimitHandle()
	if err == nil {
		handle, err = applyProcessLimits(cmd, s.engine.cfg.Execution.Resources)
		if err != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
	if err == nil {
		err = cmd.Wait()
	}
	limitStats := handle.Stats()
	handle.Cleanup()
	text := output.String()
	if max := s.engine.cfg.Reports.MaxOutputBytes; max > 0 && len(text) > max {
		text = text[:max]
	}
	status := StatusKilled
	failureKind := ""
	reason := "tests failed with mutant applied"
	if memoryLimitExceeded(err, cmd.ProcessState, s.engine.cfg.Execution.Resources, text) {
		status = StatusMemoryKilled
		failureKind = "memory_limit_exceeded"
		reason = "test process exceeded the configured memory limit"
	} else if runCtx.Err() == context.DeadlineExceeded {
		status = StatusTimedOut
		failureKind = "timeout"
		reason = "test command timed out"
	} else if err == nil {
		status = StatusSurvived
		reason = "tests passed with mutant applied"
	} else if errors.Is(err, errProcessLimitUnsupported) {
		status = StatusSkippedResource
		failureKind = "resource_limit_unsupported"
		reason = "configured process resource limits are not supported on this platform"
	} else if !strings.Contains(text, "FAIL") {
		status = StatusCompileError
		failureKind = classifyFailure(text, err)
		reason = "test command failed before running assertions"
	}
	return MutantResult{
		MutantID:        job.Mutant.ID,
		Status:          status,
		FailureKind:     failureKind,
		MemoryPeakBytes: limitStats.PeakProcessMemoryBytes,
		Duration:        time.Since(start),
		TestCommand:     job.TestCommand,
		StatusReason:    reason,
		Output:          text,
		Mutant:          job.Mutant,
	}, nil
}

func classifyFailure(output string, err error) string {
	text := strings.ToLower(output)
	switch {
	case strings.Contains(text, "panic:"):
		return "test_panic"
	case strings.Contains(text, "build failed"), strings.Contains(text, "compilation failed"), strings.Contains(text, "undefined:"), strings.Contains(text, "syntax error"):
		return "compile_error"
	case strings.Contains(text, "no such file or directory"), strings.Contains(text, "cannot find"):
		return "environment_error"
	case err != nil:
		return "runner_error"
	default:
		return ""
	}
}
