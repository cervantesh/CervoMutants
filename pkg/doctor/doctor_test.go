package doctor

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestWarningChecksDoNotFailDoctor(t *testing.T) {
	check := warning("windows-onedrive", "workspace is under OneDrive")
	if !check.OK {
		t.Fatal("warning checks should not fail doctor")
	}
	if check.Severity != "warn" {
		t.Fatalf("severity = %q, want warn", check.Severity)
	}
}

func TestRunIncludesCommandAndRuntimeChecks(t *testing.T) {
	checks := Run(context.Background())
	if len(checks) < 2 {
		t.Fatalf("checks = %d, want at least command checks", len(checks))
	}
	if !containsCheck(checks, "go") || !containsCheck(checks, "git") || !containsCheck(checks, "runtime") {
		t.Fatalf("expected go/git/runtime checks: %+v", checks)
	}
}

func TestCheckCommandClassifiesMissingCommand(t *testing.T) {
	check := checkCommand(context.Background(), "cervomut-command-that-does-not-exist")
	if check.OK || check.Severity != "fail" {
		t.Fatalf("missing command check = %+v, want failed", check)
	}
}

func TestRuntimeEnvironmentHelpers(t *testing.T) {
	if !mentionsOneDrive(`C:\Users\me\OneDrive\Documents`) {
		t.Fatal("mentionsOneDrive should be case-insensitive")
	}
	if mentionsOneDrive(`/tmp/work`) {
		t.Fatal("mentionsOneDrive should ignore unrelated paths")
	}

	checks := windowsChecks(strings.Repeat("x", 121), `C:\Temp`)
	if len(checks) == 0 {
		t.Fatal("windowsChecks should report conservative Windows guidance")
	}
	checks = windowsChecks(`C:\Users\me\OneDrive\Documents\project`, `C:\Users\me\OneDrive\Temp`)
	if !containsCheck(checks, "windows-onedrive") || !containsCheck(checks, "windows-temp-onedrive") {
		t.Fatalf("windowsChecks missing OneDrive warnings: %+v", checks)
	}
	checks = windowsChecks(`\\server\share\project`, `C:\Temp`)
	if !containsCheck(checks, "windows-network-path") {
		t.Fatalf("windowsChecks missing network path warning: %+v", checks)
	}
	if runtime.GOOS != "linux" && isWSL() {
		t.Fatal("isWSL should be false outside Linux")
	}
	_ = linuxChecks()
}

func containsCheck(checks []Check, name string) bool {
	for _, check := range checks {
		if check.Name == name {
			return true
		}
	}
	return false
}
