package isolate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const markerFile = ".cervomut-workdir"

func CopyModule(moduleDir string) (string, error) {
	abs, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp("", "cervomut-"+safePathToken(abs)+"-*")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(tmp, markerFile), []byte("cervomut isolated workdir\n"), 0o600); err != nil {
		_ = os.RemoveAll(tmp)
		return "", err
	}
	if err := copyTree(abs, tmp); err != nil {
		_ = Cleanup(tmp)
		return "", err
	}
	return tmp, nil
}

func Cleanup(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(filepath.Join(path, markerFile)); err != nil {
		return errors.New("refusing to cleanup unmarked cervomut workdir")
	}
	return os.RemoveAll(path)
}

func ContainedTargetPath(moduleDir, workdir, file string) (string, error) {
	absModule, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absModule, absFile)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) || filepath.IsAbs(rel) {
		return "", errors.New("mutant file is outside module")
	}
	return filepath.Join(workdir, rel), nil
}

func safePathToken(path string) string {
	cleaned := strings.TrimSpace(filepath.Clean(path))
	normalized := strings.ReplaceAll(cleaned, "\\", "/")
	parts := strings.Split(normalized, "/")
	base := ""
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.TrimSpace(parts[i]) != "" {
			base = parts[i]
			break
		}
	}
	if base == "" {
		base = "module"
	}
	var b strings.Builder
	lastUnderscore := false
	for _, r := range base {
		allowed := r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_')
		if allowed {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	token := strings.Trim(b.String(), "._-")
	if token == "" {
		token = "module"
	}
	if len(token) > 48 {
		token = token[:48]
		token = strings.TrimRight(token, "._-")
	}
	sum := sha256.Sum256([]byte(cleaned))
	return token + "-" + hex.EncodeToString(sum[:])[:12]
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() && excludedDir(d.Name()) {
			return filepath.SkipDir
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func excludedDir(name string) bool {
	switch name {
	case ".git", ".cervomut", "vendor", "node_modules", "dist", "build", "coverage":
		return true
	default:
		return false
	}
}
