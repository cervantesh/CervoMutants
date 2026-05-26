package isolate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSafePathTokenRejectsWindowsInvalidFilenameCharacters(t *testing.T) {
	token := safePathToken(`C:\Users\c___h\OneDrive\Documents\CervoSoft\cobra doc`)

	for _, invalid := range []string{":", "\\", "/", "*", "?", "\"", "<", ">", "|", " "} {
		if strings.Contains(token, invalid) {
			t.Fatalf("safe token %q contains invalid filename character %q", token, invalid)
		}
	}
	if !strings.HasPrefix(token, "cobra_doc-") {
		t.Fatalf("token = %q, want basename plus hash", token)
	}
	if len(token) > 80 {
		t.Fatalf("token length = %d, want bounded token: %q", len(token), token)
	}
}

func TestCopyModuleUsesSafeWorkdirNameForWindowsStylePaths(t *testing.T) {
	module := t.TempDir()
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "main.go"), []byte("package fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	workdir, err := CopyModule(module)
	if err != nil {
		t.Fatalf("CopyModule returned error: %v", err)
	}
	defer Cleanup(workdir)

	base := filepath.Base(workdir)
	if !strings.HasPrefix(base, "cervomut-") {
		t.Fatalf("workdir base = %q, want cervomut prefix", base)
	}
	for _, invalid := range []string{":", "*", "?", "\"", "<", ">", "|"} {
		if strings.Contains(base, invalid) {
			t.Fatalf("workdir base %q contains invalid filename character %q", base, invalid)
		}
	}
	if runtime.GOOS == "windows" && strings.Contains(base, "\\") {
		t.Fatalf("workdir base %q contains path separator", base)
	}
	if _, err := os.Stat(filepath.Join(workdir, ".cervomut-workdir")); err != nil {
		t.Fatalf("workdir marker missing: %v", err)
	}
}

func TestCleanupRefusesUnmarkedPath(t *testing.T) {
	dir := t.TempDir()
	if err := Cleanup(dir); err == nil {
		t.Fatal("Cleanup succeeded for unmarked path")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("unmarked path was removed or changed: %v", err)
	}
}

func TestContainedTargetPathRejectsFileOutsideModule(t *testing.T) {
	root := t.TempDir()
	module := filepath.Join(root, "module")
	workdir := filepath.Join(root, "workdir")
	outside := filepath.Join(root, "outside.go")
	if err := os.MkdirAll(module, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workdir, 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := ContainedTargetPath(module, workdir, outside); err == nil {
		t.Fatal("ContainedTargetPath accepted a file outside the module")
	}
}

func TestContainedTargetPathMapsModuleFileIntoWorkdir(t *testing.T) {
	root := t.TempDir()
	module := filepath.Join(root, "module")
	workdir := filepath.Join(root, "workdir")
	file := filepath.Join(module, "pkg", "calc.go")
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(workdir, 0o700); err != nil {
		t.Fatal(err)
	}

	target, err := ContainedTargetPath(module, workdir, file)
	if err != nil {
		t.Fatalf("ContainedTargetPath returned error: %v", err)
	}
	want := filepath.Join(workdir, "pkg", "calc.go")
	if target != want {
		t.Fatalf("target = %q, want %q", target, want)
	}
}
