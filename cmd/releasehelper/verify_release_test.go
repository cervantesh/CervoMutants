package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdVerifyReleaseAcceptsAlignedArtifacts(t *testing.T) {
	dir := t.TempDir()
	dist := filepath.Join(dir, "dist")
	upgradeDir := filepath.Join(dir, "docs", "upgrade-notes")
	version := "v9.9.9"

	writeFile(t, filepath.Join(upgradeDir, version+".md"), `# Upgrade Notes for v9.9.9

## Summary

- Example release summary.

## Operator Action

- Read the release notes before upgrading.

## Rollback

- Reinstall the previous known-good version if validation fails.
`)
	writeFile(t, filepath.Join(dist, "release-notes.md"), "# CervoMutants v9.9.9\n\n## Changelog\n\n- example\n\n## Upgrade Notes\n\n## Summary\n\n- Example release summary.\n\n## Operator Action\n\n- Read the release notes before upgrading.\n\n## Rollback\n\n- Reinstall the previous known-good version if validation fails.\n")
	writeReleaseAlignmentFixtures(t, dir, version)

	var checksumLines []string
	for _, target := range supportedReleaseTargets() {
		archivePath := filepath.Join(dist, target.archiveName(version))
		writeReleaseArchive(t, archivePath, target, version)
		sum := fileHashForTest(t, archivePath)
		checksumLines = append(checksumLines, sum+"  "+target.archiveName(version))
	}
	writeFile(t, filepath.Join(dist, "SHA256SUMS"), strings.Join(checksumLines, "\n")+"\n")

	manifestPath := filepath.Join(dist, "release-manifest.json")
	err := cmdVerifyRelease([]string{
		"--version", version,
		"--dist", dist,
		"--notes", filepath.Join(dist, "release-notes.md"),
		"--upgrade-dir", upgradeDir,
		"--manifest-out", manifestPath,
		"--repo-root", dir,
	})
	if err != nil {
		t.Fatalf("cmdVerifyRelease returned error: %v", err)
	}
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
}

func TestCmdVerifyReleaseRejectsMissingRollbackSection(t *testing.T) {
	dir := t.TempDir()
	version := "v9.9.9"
	upgradeDir := filepath.Join(dir, "docs", "upgrade-notes")
	writeFile(t, filepath.Join(upgradeDir, version+".md"), `# Upgrade Notes for v9.9.9

## Summary

- Example release summary.

## Operator Action

- Read the release notes before upgrading.
`)

	err := cmdVerifyRelease([]string{
		"--version", version,
		"--dist", filepath.Join(dir, "dist"),
		"--notes", filepath.Join(dir, "dist", "release-notes.md"),
		"--upgrade-dir", upgradeDir,
	})
	if err == nil || !strings.Contains(err.Error(), "## Rollback") {
		t.Fatalf("expected rollback-section error, got %v", err)
	}
}

func TestCmdVerifyReleaseRejectsStaleExternalWaveDefault(t *testing.T) {
	dir := t.TempDir()
	version := "v9.9.9"
	dist := filepath.Join(dir, "dist")
	upgradeDir := filepath.Join(dir, "docs", "upgrade-notes")
	writeFile(t, filepath.Join(upgradeDir, version+".md"), `# Upgrade Notes for v9.9.9

## Summary

- Example release summary.

## Operator Action

- Read the release notes before upgrading.

## Rollback

- Reinstall the previous known-good version if validation fails.
`)
	writeFile(t, filepath.Join(dist, "release-notes.md"), "# CervoMutants v9.9.9\n\n## Changelog\n\n- example\n\n## Upgrade Notes\n\n## Summary\n\n- Example release summary.\n\n## Operator Action\n\n- Read the release notes before upgrading.\n\n## Rollback\n\n- Reinstall the previous known-good version if validation fails.\n")
	writeReleaseAlignmentFixtures(t, dir, version)
	writeFile(t, filepath.Join(dir, ".github", "workflows", "external-action-wave.yml"), `name: external-action-wave

on:
  workflow_dispatch:
    inputs:
      manifest_path:
        default: docs/evaluations/external-github-action-wave-v9.9.8-candidates.json
`)

	err := cmdVerifyRelease([]string{
		"--version", version,
		"--dist", dist,
		"--notes", filepath.Join(dist, "release-notes.md"),
		"--upgrade-dir", upgradeDir,
		"--repo-root", dir,
	})
	if err == nil || !strings.Contains(err.Error(), "workflow_dispatch manifest_path default") {
		t.Fatalf("expected stale workflow-default error, got %v", err)
	}
}

func TestCmdVerifyReleaseRejectsLaggingVersionPinnedDocs(t *testing.T) {
	dir := t.TempDir()
	version := "v9.9.9"
	dist := filepath.Join(dir, "dist")
	upgradeDir := filepath.Join(dir, "docs", "upgrade-notes")
	writeFile(t, filepath.Join(upgradeDir, version+".md"), `# Upgrade Notes for v9.9.9

## Summary

- Example release summary.

## Operator Action

- Read the release notes before upgrading.

## Rollback

- Reinstall the previous known-good version if validation fails.
`)
	writeFile(t, filepath.Join(dist, "release-notes.md"), "# CervoMutants v9.9.9\n\n## Changelog\n\n- example\n\n## Upgrade Notes\n\n## Summary\n\n- Example release summary.\n\n## Operator Action\n\n- Read the release notes before upgrading.\n\n## Rollback\n\n- Reinstall the previous known-good version if validation fails.\n")
	writeReleaseAlignmentFixtures(t, dir, version)
	writeFile(t, filepath.Join(dir, "docs", "install.md"), "go install github.com/cervantesh/cervo-mutants/cmd/cervomut@v9.9.8\n")

	err := cmdVerifyRelease([]string{
		"--version", version,
		"--dist", dist,
		"--notes", filepath.Join(dist, "release-notes.md"),
		"--upgrade-dir", upgradeDir,
		"--repo-root", dir,
	})
	if err == nil || !strings.Contains(err.Error(), "docs/install.md references release") {
		t.Fatalf("expected lagging version-pinned doc error, got %v", err)
	}
}

func writeReleaseAlignmentFixtures(t *testing.T, root, version string) {
	t.Helper()
	writeFile(t, filepath.Join(root, ".github", "workflows", "external-action-wave.yml"), fmt.Sprintf(`name: external-action-wave

on:
  workflow_dispatch:
    inputs:
      manifest_path:
        default: docs/evaluations/external-github-action-wave-%s-candidates.json
`, version))
	writeFile(t, filepath.Join(root, "docs", "evaluations", fmt.Sprintf("external-github-action-wave-%s-candidates.json", version)), fmt.Sprintf(`{
  "install_path": "github-action@%s",
  "action_ref": "%s"
}
`, version, version))
	writeFile(t, filepath.Join(root, "docs", "install.md"), fmt.Sprintf("go install github.com/cervantesh/cervo-mutants/cmd/cervomut@%s\nversion=%s\n", version, version))
	writeFile(t, filepath.Join(root, "docs", "github-action.md"), fmt.Sprintf("uses: cervantesh/cervo-mutants@%s\nrefs/tags/%s\n", version, version))
	writeFile(t, filepath.Join(root, "docs", "adoption-guide.md"), fmt.Sprintf("latest released wave against %s\n", version))
	writeFile(t, filepath.Join(root, "docs", "rollout-playbooks.md"), fmt.Sprintf("current released hosted evidence %s\n", version))
	writeFile(t, filepath.Join(root, ".github", "ISSUE_TEMPLATE", "adoption-feedback.yml"), fmt.Sprintf("placeholder: %s\n", version))
}

func writeReleaseArchive(t *testing.T, archivePath string, target releaseTarget, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		t.Fatal(err)
	}
	stageRoot := target.stageRoot(version)
	files := expectedStageFiles(target)
	switch target.Format {
	case "zip":
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		zw := zip.NewWriter(f)
		for _, name := range files {
			w, err := zw.Create(filepath.ToSlash(filepath.Join(stageRoot, name)))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte(name)); err != nil {
				t.Fatal(err)
			}
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	case "tar.gz":
		f, err := os.Create(archivePath)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		gzw := gzip.NewWriter(f)
		tw := tar.NewWriter(gzw)
		for _, name := range files {
			full := filepath.ToSlash(filepath.Join(stageRoot, name))
			body := []byte(name)
			hdr := &tar.Header{Name: full, Mode: 0o644, Size: int64(len(body))}
			if err := tw.WriteHeader(hdr); err != nil {
				t.Fatal(err)
			}
			if _, err := tw.Write(body); err != nil {
				t.Fatal(err)
			}
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gzw.Close(); err != nil {
			t.Fatal(err)
		}
	default:
		t.Fatalf("unsupported archive format %q", target.Format)
	}
}

func fileHashForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
