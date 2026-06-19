package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
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
