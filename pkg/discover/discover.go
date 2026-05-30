package discover

import (
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	ModuleDir string `json:"module_dir"`
	Package   string `json:"package"`
	Path      string `json:"path"`
	IsTest    bool   `json:"is_test"`
}

type Result struct {
	Modules []string `json:"modules"`
	Files   []File   `json:"files"`
}

func Discover(targets []string) (Result, error) {
	if len(targets) == 0 {
		targets = []string{"."}
	}
	var result Result
	seenModules := map[string]bool{}
	for _, target := range targets {
		walkRoot, err := walkRootForTarget(target)
		if err != nil {
			return Result{}, err
		}
		moduleDir := moduleDirForWalkRoot(walkRoot)
		if !seenModules[moduleDir] {
			seenModules[moduleDir] = true
			result.Modules = append(result.Modules, moduleDir)
		}
		if err := appendWalkFiles(&result, walkRoot, moduleDir); err != nil {
			return Result{}, err
		}
	}
	return result, nil
}

func walkRootForTarget(target string) (string, error) {
	root := strings.TrimSuffix(target, "/...")
	if root == "." || root == "./..." {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	walkRoot := abs
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		walkRoot = resolved
	}
	if info, err := os.Stat(walkRoot); err == nil && info.IsDir() && !strings.HasSuffix(walkRoot, string(os.PathSeparator)) {
		walkRoot += string(os.PathSeparator)
	}
	return walkRoot, nil
}

func moduleDirForWalkRoot(walkRoot string) string {
	moduleDir := findModule(walkRoot)
	if moduleDir == "" {
		moduleDir = walkRoot
	}
	return filepath.Clean(moduleDir)
}

func appendWalkFiles(result *Result, walkRoot, moduleDir string) error {
	return filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return handleDir(d)
		}
		file, ok := discoveredGoFile(moduleDir, path)
		if !ok {
			return nil
		}
		result.Files = append(result.Files, file)
		return nil
	})
}

func handleDir(d os.DirEntry) error {
	if excludedDir(d.Name()) {
		return filepath.SkipDir
	}
	return nil
}

func discoveredGoFile(moduleDir, path string) (File, bool) {
	if !strings.HasSuffix(path, ".go") {
		return File{}, false
	}
	name := filepath.Base(path)
	if generatedFile(name) {
		return File{}, false
	}
	rel, _ := filepath.Rel(moduleDir, filepath.Dir(path))
	pkg := "./" + filepath.ToSlash(rel)
	if rel == "." {
		pkg = "."
	}
	return File{ModuleDir: moduleDir, Package: pkg, Path: path, IsTest: strings.HasSuffix(name, "_test.go")}, true
}

func generatedFile(name string) bool {
	return strings.HasSuffix(name, "_generated.go") || strings.HasSuffix(name, ".pb.go")
}

func findModule(start string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func excludedDir(name string) bool {
	switch name {
	case ".git", ".cervomut", "vendor", "node_modules", "dist", "build", "coverage":
		return true
	default:
		return false
	}
}
