package pool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RealCommandRunner struct {
	Monitor MemoryMonitor
}

func (r RealCommandRunner) Run(ctx context.Context, spec CommandSpec) (CommandResult, error) {
	if spec.Path == "" {
		return CommandResult{}, errors.New("command path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(spec.LogPath), 0o755); err != nil {
		return CommandResult{}, err
	}
	monitor := r.Monitor
	if monitor == nil {
		monitor = systemMemoryMonitor{}
	}
	if err := waitFreeMemory(monitor, spec); err != nil {
		if writeErr := os.WriteFile(spec.LogPath, []byte(err.Error()), 0o644); writeErr != nil {
			return CommandResult{}, writeErr
		}
		return CommandResult{ExitCode: 125}, nil
	}

	stdoutPath := spec.LogPath + ".stdout"
	stderrPath := spec.LogPath + ".stderr"
	_ = os.Remove(stdoutPath)
	_ = os.Remove(stderrPath)
	_ = os.Remove(spec.LogPath)
	stdout, err := os.Create(stdoutPath)
	if err != nil {
		return CommandResult{}, err
	}
	defer stdout.Close()
	stderr, err := os.Create(stderrPath)
	if err != nil {
		return CommandResult{}, err
	}
	defer stderr.Close()

	cmd := exec.Command(spec.Path, spec.Args...)
	cmd.Dir = spec.Dir
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		text := err.Error()
		if writeErr := os.WriteFile(spec.LogPath, []byte(text), 0o644); writeErr != nil {
			return CommandResult{}, writeErr
		}
		return CommandResult{ExitCode: 125}, nil
	}

	procMonitor, err := attachProcessMonitor(cmd)
	if err != nil && spec.MaxProcessTreeMemoryMB > 0 {
		killCommand(cmd, procMonitor)
		text := err.Error()
		if writeErr := os.WriteFile(spec.LogPath, []byte(text), 0o644); writeErr != nil {
			return CommandResult{}, writeErr
		}
		return CommandResult{ExitCode: 125}, nil
	}
	defer procMonitor.close()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	pollEvery := spec.MemoryPoll
	if pollEvery <= 0 {
		pollEvery = 5 * time.Second
	}
	ticker := time.NewTicker(pollEvery)
	defer ticker.Stop()

	timeout := spec.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	runStarted := time.Now()
	deadline := runStarted.Add(timeout)
	limitBytes := int64(spec.MaxProcessTreeMemoryMB) * 1024 * 1024

	for {
		select {
		case err := <-waitCh:
			exitCode := 0
			if err != nil && cmd.ProcessState != nil {
				exitCode = cmd.ProcessState.ExitCode()
			}
			if writeErr := stitchLogs(stdoutPath, stderrPath, spec.LogPath); writeErr != nil {
				return CommandResult{}, writeErr
			}
			return CommandResult{ExitCode: exitCode}, nil
		case <-ticker.C:
			if time.Now().After(deadline) {
				killCommand(cmd, procMonitor)
				<-waitCh
				msg := fmt.Sprintf("timed out after %ds", int(timeout.Seconds()))
				if writeErr := os.WriteFile(spec.LogPath, []byte(msg), 0o644); writeErr != nil {
					return CommandResult{}, writeErr
				}
				return CommandResult{ExitCode: 124}, nil
			}
			if spec.KillBelowFreeMemoryMB > 0 || spec.KillBelowFreeCommitMB > 0 {
				status, statusErr := monitor.Status()
				if statusErr == nil {
					belowMemory := spec.KillBelowFreeMemoryMB > 0 && status.FreeMemoryMB < spec.KillBelowFreeMemoryMB
					belowCommit := spec.KillBelowFreeCommitMB > 0 && status.FreeCommitMB < spec.KillBelowFreeCommitMB
					if belowMemory || belowCommit {
						killCommand(cmd, procMonitor)
						<-waitCh
						msg := fmt.Sprintf("killed by memory watchdog after %ds; free memory=%dMB free commit=%dMB", int(time.Since(runStarted).Seconds()), status.FreeMemoryMB, status.FreeCommitMB)
						if writeErr := os.WriteFile(spec.LogPath, []byte(msg), 0o644); writeErr != nil {
							return CommandResult{}, writeErr
						}
						return CommandResult{ExitCode: 126}, nil
					}
				}
			}
			if limitBytes > 0 {
				peak := procMonitor.peakJobMemoryBytes()
				if peak > limitBytes {
					killCommand(cmd, procMonitor)
					<-waitCh
					msg := fmt.Sprintf("killed by process-tree memory watchdog; peak job memory=%dMB limit=%dMB", peak/(1024*1024), spec.MaxProcessTreeMemoryMB)
					if writeErr := os.WriteFile(spec.LogPath, []byte(msg), 0o644); writeErr != nil {
						return CommandResult{}, writeErr
					}
					return CommandResult{ExitCode: 126}, nil
				}
			}
		case <-ctx.Done():
			killCommand(cmd, procMonitor)
			<-waitCh
			msg := "command canceled"
			if writeErr := os.WriteFile(spec.LogPath, []byte(msg), 0o644); writeErr != nil {
				return CommandResult{}, writeErr
			}
			return CommandResult{ExitCode: 124}, nil
		}
	}
}

func waitFreeMemory(monitor MemoryMonitor, spec CommandSpec) error {
	if (spec.MinFreeMemoryMB <= 0 && spec.MinFreeCommitMB <= 0) || monitor == nil {
		return nil
	}
	timeout := spec.MemoryWait
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	for {
		status, err := monitor.Status()
		if err == nil {
			hasMemory := spec.MinFreeMemoryMB <= 0 || status.FreeMemoryMB >= spec.MinFreeMemoryMB
			hasCommit := spec.MinFreeCommitMB <= 0 || status.FreeCommitMB >= spec.MinFreeCommitMB
			if hasMemory && hasCommit {
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("skipped after waiting %ds for %dMB free memory and %dMB free commit; free memory=%dMB free commit=%dMB", int(timeout.Seconds()), spec.MinFreeMemoryMB, spec.MinFreeCommitMB, status.FreeMemoryMB, status.FreeCommitMB)
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func killCommand(cmd *exec.Cmd, monitor processMonitor) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	monitor.close()
}

func stitchLogs(stdoutPath, stderrPath, logPath string) error {
	parts := make([]string, 0, 2)
	if data, err := os.ReadFile(stdoutPath); err == nil && len(data) > 0 {
		parts = append(parts, string(data))
	}
	if data, err := os.ReadFile(stderrPath); err == nil && len(data) > 0 {
		parts = append(parts, string(data))
	}
	_ = os.Remove(stdoutPath)
	_ = os.Remove(stderrPath)
	return os.WriteFile(logPath, []byte(strings.Join(parts, "\n")), 0o644)
}
