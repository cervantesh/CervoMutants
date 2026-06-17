package isolate

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type TempRootPlan struct {
	Root     string
	Source   string
	Warnings []string
}

func ResolveTempRoot(moduleDir, configured string) TempRootPlan {
	configured = strings.TrimSpace(configured)
	if configured != "" {
		return TempRootPlan{
			Root:     filepath.Clean(configured),
			Source:   "configured",
			Warnings: windowsTempRootWarnings(moduleDir, os.TempDir()),
		}
	}
	if runtime.GOOS != "windows" {
		return TempRootPlan{Root: os.TempDir(), Source: "system"}
	}
	return resolveWindowsTempRoot(moduleDir, os.TempDir(), os.Getenv("LOCALAPPDATA"), os.Getenv("USERPROFILE"), os.Getenv("SystemDrive"))
}

func CreateTempDir(moduleDir, configured, pattern string) (string, error) {
	plan := ResolveTempRoot(moduleDir, configured)
	if plan.Root != "" {
		if err := os.MkdirAll(plan.Root, 0o755); err == nil {
			return os.MkdirTemp(plan.Root, pattern)
		}
	}
	return os.MkdirTemp("", pattern)
}

func resolveWindowsTempRoot(moduleDir, systemTemp, localAppData, userProfile, systemDrive string) TempRootPlan {
	plan := TempRootPlan{
		Root:     filepath.Clean(systemTemp),
		Source:   "system",
		Warnings: windowsTempRootWarnings(moduleDir, systemTemp),
	}
	if !windowsTempRootRisky(moduleDir, systemTemp) {
		return plan
	}
	for _, candidate := range windowsTempRootCandidates(localAppData, userProfile, systemDrive) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		candidate = filepath.Clean(candidate)
		if mentionsOneDrive(candidate) || isWindowsUNCPath(candidate) || pathContainsOther(candidate, moduleDir) {
			continue
		}
		plan.Root = candidate
		plan.Source = "windows-local-fallback"
		return plan
	}
	plan.Source = "system-fallback"
	return plan
}

func windowsTempRootRisky(moduleDir, systemTemp string) bool {
	return mentionsOneDrive(systemTemp) || isWindowsUNCPath(systemTemp) || pathContainsOther(moduleDir, systemTemp)
}

func windowsTempRootWarnings(moduleDir, systemTemp string) []string {
	var warnings []string
	if mentionsOneDrive(moduleDir) {
		warnings = append(warnings, "workspace path is under OneDrive")
	}
	if mentionsOneDrive(systemTemp) {
		warnings = append(warnings, "system temp path is under OneDrive")
	}
	if isWindowsUNCPath(moduleDir) {
		warnings = append(warnings, "workspace path is on a network/UNC path")
	}
	if pathContainsOther(moduleDir, systemTemp) {
		warnings = append(warnings, "system temp path shares the workspace tree")
	}
	if strings.Contains(moduleDir, " ") {
		warnings = append(warnings, "workspace path contains spaces")
	}
	return warnings
}

func windowsTempRootCandidates(localAppData, userProfile, systemDrive string) []string {
	var candidates []string
	if localAppData = strings.TrimSpace(localAppData); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "CervoMutants", "tmp"))
	}
	if userProfile = strings.TrimSpace(userProfile); userProfile != "" {
		candidates = append(candidates, filepath.Join(userProfile, "AppData", "Local", "CervoMutants", "tmp"))
	}
	if root := driveRootPath(systemDrive); root != "" {
		candidates = append(candidates, filepath.Join(root, "cervomut-tmp"))
	}
	return candidates
}

func driveRootPath(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[1] == ':' {
		return value + `\`
	}
	return ""
}

func mentionsOneDrive(path string) bool {
	return strings.Contains(strings.ToLower(path), "onedrive")
}

func isWindowsUNCPath(path string) bool {
	return strings.HasPrefix(strings.TrimSpace(path), `\\`)
}

func pathContainsOther(a, b string) bool {
	left := comparableWindowsPath(a)
	right := comparableWindowsPath(b)
	if left == "" || right == "" {
		return false
	}
	return left == right || strings.HasPrefix(left, right+`\`) || strings.HasPrefix(right, left+`\`)
}

func comparableWindowsPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	path = filepath.Clean(path)
	path = strings.ReplaceAll(path, "/", `\`)
	path = strings.TrimRight(path, `\`)
	return strings.ToLower(path)
}
