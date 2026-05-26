package discover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFollowsRootSymlink(t *testing.T) {
	dir := t.TempDir()
	module := filepath.Join(dir, "module")
	if err := os.MkdirAll(module, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module fixture\n\ngo 1.25.6\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "calc.go"), []byte("package fixture\n\nfunc Positive(n int) bool { return n > 0 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "linked-module")
	if err := os.Symlink(module, link); err != nil {
		t.Skipf("symlink unavailable on this filesystem: %v", err)
	}

	result, err := Discover([]string{link})
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("files = %d, want 1: %+v", len(result.Files), result)
	}
	if result.Files[0].Package != "." {
		t.Fatalf("package = %q, want .", result.Files[0].Package)
	}
}

func TestDiscoverHandlesRootWithTrailingSeparator(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module fixture\n\ngo 1.25.6\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc.go"), []byte("package fixture\n\nfunc Positive(n int) bool { return n > 0 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := Discover([]string{dir + string(os.PathSeparator)})
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("files = %d, want 1: %+v", len(result.Files), result)
	}
}
