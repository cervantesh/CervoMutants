package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type releaseTarget struct {
	GOOS   string
	GOARCH string
	Format string
}

type releaseManifest struct {
	Version          string                 `json:"version"`
	ReleaseNotesPath string                 `json:"release_notes_path"`
	UpgradeNotePath  string                 `json:"upgrade_note_path"`
	Assets           []releaseAssetManifest `json:"assets"`
}

type releaseAssetManifest struct {
	FileName  string   `json:"file_name"`
	GOOS      string   `json:"goos"`
	GOARCH    string   `json:"goarch"`
	Format    string   `json:"format"`
	SHA256    string   `json:"sha256"`
	StageRoot string   `json:"stage_root"`
	Files     []string `json:"files"`
}

type archiveContents struct {
	StageRoot string
	Files     []string
}

func cmdVerifyRelease(args []string) error {
	fs := flag.NewFlagSet("verify-release", flag.ContinueOnError)
	version := fs.String("version", "", "release version such as v0.3.0")
	distDir := fs.String("dist", "dist", "directory containing built release artifacts")
	notesPath := fs.String("notes", filepath.Join("dist", "release-notes.md"), "path to generated release notes")
	upgradeDir := fs.String("upgrade-dir", filepath.Join("docs", "upgrade-notes"), "directory containing per-version upgrade notes")
	manifestOut := fs.String("manifest-out", "", "optional JSON manifest output path")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*version) == "" {
		return fmt.Errorf("verify-release requires --version")
	}

	upgradePath := filepath.Join(*upgradeDir, *version+".md")
	if err := verifyUpgradeNoteStructure(upgradePath); err != nil {
		return err
	}
	if err := verifyReleaseNotes(*notesPath, *version); err != nil {
		return err
	}

	checksums, err := loadChecksums(filepath.Join(*distDir, "SHA256SUMS"))
	if err != nil {
		return err
	}

	targets := supportedReleaseTargets()
	manifest := releaseManifest{
		Version:          *version,
		ReleaseNotesPath: filepath.ToSlash(*notesPath),
		UpgradeNotePath:  filepath.ToSlash(upgradePath),
	}

	for _, target := range targets {
		archiveName := target.archiveName(*version)
		archivePath := filepath.Join(*distDir, archiveName)
		if _, err := os.Stat(archivePath); err != nil {
			return fmt.Errorf("expected release archive %s: %w", filepath.ToSlash(archivePath), err)
		}
		wantChecksum, ok := checksums[archiveName]
		if !ok {
			return fmt.Errorf("SHA256SUMS is missing %s", archiveName)
		}
		gotChecksum, err := fileSHA256(archivePath)
		if err != nil {
			return fmt.Errorf("hash %s: %w", filepath.ToSlash(archivePath), err)
		}
		if gotChecksum != wantChecksum {
			return fmt.Errorf("checksum mismatch for %s: got %s want %s", archiveName, gotChecksum, wantChecksum)
		}
		contents, err := verifyArchiveContents(archivePath, target, *version)
		if err != nil {
			return err
		}
		manifest.Assets = append(manifest.Assets, releaseAssetManifest{
			FileName:  archiveName,
			GOOS:      target.GOOS,
			GOARCH:    target.GOARCH,
			Format:    target.Format,
			SHA256:    gotChecksum,
			StageRoot: contents.StageRoot,
			Files:     contents.Files,
		})
	}

	if strings.TrimSpace(*manifestOut) != "" {
		if err := os.MkdirAll(filepath.Dir(*manifestOut), 0o755); err != nil {
			return err
		}
		data, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*manifestOut, append(data, '\n'), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func supportedReleaseTargets() []releaseTarget {
	return []releaseTarget{
		{GOOS: "linux", GOARCH: "amd64", Format: "tar.gz"},
		{GOOS: "linux", GOARCH: "arm64", Format: "tar.gz"},
		{GOOS: "darwin", GOARCH: "amd64", Format: "tar.gz"},
		{GOOS: "darwin", GOARCH: "arm64", Format: "tar.gz"},
		{GOOS: "windows", GOARCH: "amd64", Format: "zip"},
	}
}

func (t releaseTarget) archiveName(version string) string {
	return fmt.Sprintf("cervomut_%s_%s_%s.%s", version, t.GOOS, t.GOARCH, t.Format)
}

func (t releaseTarget) stageRoot(version string) string {
	return fmt.Sprintf("cervomut_%s_%s_%s", version, t.GOOS, t.GOARCH)
}

func (t releaseTarget) binaryName() string {
	if t.GOOS == "windows" {
		return "cervomut.exe"
	}
	return "cervomut"
}

func expectedStageFiles(t releaseTarget) []string {
	return []string{
		t.binaryName(),
		"LICENSE",
		"NOTICE",
		"TRADEMARKS.md",
		"README.md",
		"CHANGELOG.md",
	}
}

func verifyUpgradeNoteStructure(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read upgrade notes %s: %w", filepath.ToSlash(path), err)
	}
	text := string(body)
	required := []string{"## Summary", "## Operator Action", "## Rollback"}
	for _, heading := range required {
		if !strings.Contains(text, heading) {
			return fmt.Errorf("%s must include %q", filepath.ToSlash(path), heading)
		}
	}
	return nil
}

func verifyReleaseNotes(path, version string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read release notes %s: %w", filepath.ToSlash(path), err)
	}
	text := string(body)
	required := []string{
		"# CervoMutants " + version,
		"## Changelog",
		"## Upgrade Notes",
		"## Rollback",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			return fmt.Errorf("%s must include %q", filepath.ToSlash(path), want)
		}
	}
	return nil
}

func loadChecksums(path string) (map[string]string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read checksums %s: %w", filepath.ToSlash(path), err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("%s is empty", filepath.ToSlash(path))
	}
	checksums := make(map[string]string, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid checksum line %q in %s", line, filepath.ToSlash(path))
		}
		name := strings.TrimPrefix(fields[1], "*")
		checksums[name] = fields[0]
	}
	return checksums, nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func verifyArchiveContents(path string, target releaseTarget, version string) (archiveContents, error) {
	var files []string
	var stageRoot string
	switch target.Format {
	case "zip":
		reader, err := zip.OpenReader(path)
		if err != nil {
			return archiveContents{}, fmt.Errorf("open zip %s: %w", filepath.ToSlash(path), err)
		}
		defer reader.Close()
		for _, file := range reader.File {
			if strings.HasSuffix(file.Name, "/") {
				continue
			}
			files = append(files, filepath.ToSlash(file.Name))
		}
	case "tar.gz":
		f, err := os.Open(path)
		if err != nil {
			return archiveContents{}, fmt.Errorf("open tarball %s: %w", filepath.ToSlash(path), err)
		}
		defer f.Close()
		gzr, err := gzip.NewReader(f)
		if err != nil {
			return archiveContents{}, fmt.Errorf("open gzip %s: %w", filepath.ToSlash(path), err)
		}
		defer gzr.Close()
		tr := tar.NewReader(gzr)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return archiveContents{}, fmt.Errorf("read tarball %s: %w", filepath.ToSlash(path), err)
			}
			if hdr.FileInfo().IsDir() {
				continue
			}
			files = append(files, filepath.ToSlash(hdr.Name))
		}
	default:
		return archiveContents{}, fmt.Errorf("unsupported archive format %q", target.Format)
	}
	if len(files) == 0 {
		return archiveContents{}, fmt.Errorf("archive %s contains no files", filepath.ToSlash(path))
	}
	stageRoot = target.stageRoot(version)
	expected := make(map[string]struct{}, len(expectedStageFiles(target)))
	for _, name := range expectedStageFiles(target) {
		expected[filepath.ToSlash(filepath.Join(stageRoot, name))] = struct{}{}
	}
	sort.Strings(files)
	for want := range expected {
		if !containsString(files, want) {
			return archiveContents{}, fmt.Errorf("archive %s is missing %s", filepath.ToSlash(path), want)
		}
	}
	return archiveContents{StageRoot: stageRoot, Files: files}, nil
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
