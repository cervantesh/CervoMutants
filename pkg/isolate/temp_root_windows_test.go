//go:build windows

package isolate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveWindowsTempRootFallsBackToLocalAppDataWhenSystemTempIsRisky(t *testing.T) {
	plan := resolveWindowsTempRoot(
		`C:\Users\me\OneDrive\Documents\repo`,
		`C:\Users\me\OneDrive\Temp`,
		`C:\Users\me\AppData\Local`,
		`C:\Users\me`,
		`C:`,
	)
	if plan.Source != "windows-local-fallback" {
		t.Fatalf("source = %q, want windows-local-fallback", plan.Source)
	}
	if !strings.Contains(strings.ToLower(plan.Root), strings.ToLower(`AppData\Local\CervoMutants\tmp`)) {
		t.Fatalf("root = %q, want LOCALAPPDATA fallback", plan.Root)
	}
	if len(plan.Warnings) == 0 {
		t.Fatal("expected warnings for risky workspace/temp roots")
	}
}

func TestResolveTempRootHonorsConfiguredRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "custom-root")
	plan := ResolveTempRoot(`C:\repo`, root)
	if plan.Root != root || plan.Source != "configured" {
		t.Fatalf("plan = %+v, want configured root %q", plan, root)
	}
}

func TestCopyModuleWithRootCreatesMarkedWorkdirUnderRequestedRoot(t *testing.T) {
	module := t.TempDir()
	root := filepath.Join(t.TempDir(), "temp-root")
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "main.go"), []byte("package fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workdir, err := CopyModuleWithRoot(module, root)
	if err != nil {
		t.Fatalf("CopyModuleWithRoot returned error: %v", err)
	}
	defer Cleanup(workdir)

	if !strings.HasPrefix(strings.ToLower(workdir), strings.ToLower(root)) {
		t.Fatalf("workdir = %q, want prefix %q", workdir, root)
	}
	if _, err := os.Stat(filepath.Join(workdir, markerFile)); err != nil {
		t.Fatalf("marker file missing: %v", err)
	}
}
