package doctor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Check struct {
	Name     string `json:"name"`
	OK       bool   `json:"ok"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message"`
}

func Run(ctx context.Context) []Check {
	checks := []Check{checkCommand(ctx, "go", "version"), checkCommand(ctx, "git", "--version")}
	checks = append(checks, checkRuntimeEnvironment()...)
	return checks
}

func checkCommand(ctx context.Context, name string, args ...string) Check {
	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	severity := "ok"
	if err != nil {
		severity = "fail"
	}
	return Check{Name: name, OK: err == nil, Severity: severity, Message: output.String()}
}

func checkRuntimeEnvironment() []Check {
	var checks []Check
	wd, _ := os.Getwd()
	temp := os.TempDir()
	checks = append(checks, Check{
		Name:     "runtime",
		OK:       true,
		Severity: "ok",
		Message:  fmt.Sprintf("%s/%s temp=%s\n", runtime.GOOS, runtime.GOARCH, temp),
	})
	if runtime.GOOS == "windows" {
		checks = append(checks, windowsChecks(wd, temp)...)
	}
	if runtime.GOOS == "linux" {
		checks = append(checks, linuxChecks()...)
	}
	return checks
}

func windowsChecks(wd, temp string) []Check {
	var checks []Check
	if mentionsOneDrive(wd) {
		checks = append(checks, warning("windows-onedrive", "workspace is under OneDrive; large mutation runs should use a short local temp/workdir outside synced folders\n"))
	}
	if mentionsOneDrive(temp) {
		checks = append(checks, warning("windows-temp-onedrive", "TEMP appears to be under OneDrive; configure a short local temp root for mutation runs\n"))
	}
	if len(wd) > 120 {
		checks = append(checks, warning("windows-long-path", fmt.Sprintf("workspace path is %d characters; long paths increase risk for external tools and temp workdirs\n", len(wd))))
	}
	volume := filepath.VolumeName(wd)
	if strings.HasPrefix(wd, `\\`) || strings.HasPrefix(volume, `\\`) {
		checks = append(checks, warning("windows-network-path", "workspace appears to be on a network/UNC path; local disk is recommended for mutation runs\n"))
	}
	checks = append(checks, warning("windows-resource-control", "for large Windows-native runs, use conservative workers and prefer Job Object/process-tree limits when available\n"))
	return checks
}

func linuxChecks() []Check {
	var checks []Check
	if isWSL() {
		checks = append(checks, Check{Name: "wsl", OK: true, Severity: "ok", Message: "running under WSL; Linux filesystem paths such as /tmp or $HOME are recommended over /mnt/c/OneDrive for large mutation runs\n"})
		if _, err := exec.LookPath("systemd-run"); err == nil {
			checks = append(checks, Check{Name: "cgroup-scope", OK: true, Severity: "ok", Message: "systemd-run is available for bounded local experiments\n"})
		} else {
			checks = append(checks, warning("cgroup-scope", "systemd-run not found; hard local cgroup limits may not be available\n"))
		}
	}
	return checks
}

func warning(name, message string) Check {
	return Check{Name: name, OK: true, Severity: "warn", Message: message}
}

func mentionsOneDrive(path string) bool {
	return strings.Contains(strings.ToLower(path), "onedrive")
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	text := strings.ToLower(string(data))
	return strings.Contains(text, "microsoft") || strings.Contains(text, "wsl")
}
