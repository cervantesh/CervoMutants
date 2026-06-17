package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractMarkdownSection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "CHANGELOG.md")
	body := `# Changelog

## [Unreleased]

## [v0.3.0] - 2026-06-17
### Added
- item one

## [v0.2.0] - 2026-06-17
- older
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	section, err := extractMarkdownSection(path, "v0.3.0")
	if err != nil {
		t.Fatalf("extractMarkdownSection returned error: %v", err)
	}
	for _, want := range []string{"### Added", "- item one"} {
		if !strings.Contains(section, want) {
			t.Fatalf("section missing %q: %s", want, section)
		}
	}
}

func TestBuildReleaseNotesStripsUpgradeHeading(t *testing.T) {
	notes := buildReleaseNotes("v0.3.0", "### Added\n- item one", "# Upgrade Notes for v0.3.0\n\n- migrate config\n")
	for _, want := range []string{"# CervoMutants v0.3.0", "## Changelog", "## Upgrade Notes", "- migrate config"} {
		if !strings.Contains(notes, want) {
			t.Fatalf("release notes missing %q:\n%s", want, notes)
		}
	}
	if strings.Contains(notes, "# Upgrade Notes for v0.3.0") {
		t.Fatalf("release notes kept top-level upgrade heading:\n%s", notes)
	}
}

func TestCmdNotesWritesOutput(t *testing.T) {
	dir := t.TempDir()
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	upgradeDir := filepath.Join(dir, "upgrade")
	if err := os.MkdirAll(upgradeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changelogPath, []byte("# Changelog\n\n## [v0.3.0] - 2026-06-17\n- note\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(upgradeDir, "v0.3.0.md"), []byte("# Upgrade Notes\n\n- migrate\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "dist", "release-notes.md")
	if err := cmdNotes([]string{"--version", "v0.3.0", "--changelog", changelogPath, "--upgrade-dir", upgradeDir, "--out", outPath}); err != nil {
		t.Fatalf("cmdNotes returned error: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read release notes: %v", err)
	}
	if !strings.Contains(string(data), "- note") || !strings.Contains(string(data), "- migrate") {
		t.Fatalf("unexpected release notes body:\n%s", data)
	}
}
