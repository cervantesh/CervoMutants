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

func TestSafePathTokenHandlesEmptyUnsafeAndLongNames(t *testing.T) {
	for _, input := range []string{"", "   ", "///", "!!!", strings.Repeat("a", 80)} {
		token := safePathToken(input)
		if token == "" || strings.HasPrefix(token, "-") || strings.HasSuffix(token, "-") {
			t.Fatalf("unsafe token for %q: %q", input, token)
		}
		if len(token) > 70 {
			t.Fatalf("token not bounded for %q: %q", input, token)
		}
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

func TestCopyModuleExcludesHeavyDirectoriesAndCopiesNestedFiles(t *testing.T) {
	module := t.TempDir()
	for _, path := range []string{
		"go.mod",
		filepath.Join("pkg", "calc.go"),
		filepath.Join("vendor", "ignored.go"),
		filepath.Join(".git", "ignored"),
		filepath.Join("node_modules", "ignored.js"),
	} {
		full := filepath.Join(module, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	workdir, err := CopyModule(module)
	if err != nil {
		t.Fatalf("CopyModule returned error: %v", err)
	}
	defer Cleanup(workdir)
	if _, err := os.Stat(filepath.Join(workdir, "pkg", "calc.go")); err != nil {
		t.Fatalf("nested file was not copied: %v", err)
	}
	for _, path := range []string{filepath.Join("vendor", "ignored.go"), filepath.Join(".git", "ignored"), filepath.Join("node_modules", "ignored.js")} {
		if _, err := os.Stat(filepath.Join(workdir, path)); !os.IsNotExist(err) {
			t.Fatalf("excluded path %s copied or returned unexpected err: %v", path, err)
		}
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

func TestCleanupAcceptsMarkedPathAndEmptyPath(t *testing.T) {
	if err := Cleanup(""); err != nil {
		t.Fatalf("empty cleanup returned error: %v", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, markerFile), []byte("marker"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Cleanup(dir); err != nil {
		t.Fatalf("Cleanup marked path returned error: %v", err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("marked path still exists or unexpected err: %v", err)
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

func TestExcludedDirMatchesHeavyOrGeneratedTrees(t *testing.T) {
	for _, name := range []string{".git", ".cervomut", "vendor", "node_modules", "dist", "build", "coverage"} {
		if !excludedDir(name) {
			t.Fatalf("excludedDir(%q) = false, want true", name)
		}
	}
	if excludedDir("pkg") {
		t.Fatal("ordinary package directory should not be excluded")
	}
}
