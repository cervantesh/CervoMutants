package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: releasehelper <notes>")
	}
	switch args[0] {
	case "notes":
		return cmdNotes(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func cmdNotes(args []string) error {
	fs := flag.NewFlagSet("notes", flag.ContinueOnError)
	version := fs.String("version", "", "release version such as v0.3.0")
	changelogPath := fs.String("changelog", "CHANGELOG.md", "path to changelog")
	upgradeDir := fs.String("upgrade-dir", filepath.Join("docs", "upgrade-notes"), "directory containing per-version upgrade notes")
	out := fs.String("out", "", "optional output path")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if strings.TrimSpace(*version) == "" {
		return fmt.Errorf("notes requires --version")
	}
	changelogSection, err := extractMarkdownSection(*changelogPath, *version)
	if err != nil {
		return err
	}
	upgradePath := filepath.Join(*upgradeDir, *version+".md")
	upgradeBody, err := os.ReadFile(upgradePath)
	if err != nil {
		return fmt.Errorf("read upgrade notes %s: %w", filepath.ToSlash(upgradePath), err)
	}
	notes := buildReleaseNotes(*version, changelogSection, string(upgradeBody))
	if strings.TrimSpace(*out) == "" {
		fmt.Print(notes)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		return err
	}
	return os.WriteFile(*out, []byte(notes), 0o644)
}

func extractMarkdownSection(path, version string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	var start int = -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## ["+version+"]" || strings.HasPrefix(trimmed, "## ["+version+"] ") || trimmed == "## "+version || strings.HasPrefix(trimmed, "## "+version+" ") {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return "", fmt.Errorf("version %s not found in %s", version, filepath.ToSlash(path))
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			end = i
			break
		}
	}
	section := strings.TrimSpace(strings.Join(lines[start:end], "\n"))
	if section == "" {
		return "", fmt.Errorf("version %s section in %s is empty", version, filepath.ToSlash(path))
	}
	return section, nil
}

func buildReleaseNotes(version, changelogSection, upgradeNotes string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# CervoMutants %s\n\n", version)
	b.WriteString("## Changelog\n\n")
	b.WriteString(strings.TrimSpace(changelogSection))
	b.WriteString("\n\n## Upgrade Notes\n\n")
	b.WriteString(stripTopHeading(strings.TrimSpace(upgradeNotes)))
	b.WriteString("\n")
	return b.String()
}

func stripTopHeading(body string) string {
	if body == "" {
		return body
	}
	lines := strings.Split(body, "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "#") {
		return strings.TrimSpace(strings.Join(lines[1:], "\n"))
	}
	return body
}
