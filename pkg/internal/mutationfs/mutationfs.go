package mutationfs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/isolate"
)

func ApplyReplacement(path string, startOffset, endOffset int, original, mutated string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if startOffset < 0 || endOffset > len(data) || startOffset >= endOffset {
		return errors.New("mutant patch offsets are invalid")
	}
	if !strings.Contains(string(data[startOffset:endOffset]), original) {
		return errors.New("original token not found at mutant patch offsets")
	}
	segment := string(data[startOffset:endOffset])
	replaced := strings.Replace(segment, original, mutated, 1)
	text := string(data[:startOffset]) + replaced + string(data[endOffset:])
	return os.WriteFile(path, []byte(text), 0o644)
}

func PrepareOverlay(module, file string, startOffset, endOffset int, original, mutated, tempRoot string) (string, string, func(), error) {
	tmp, err := isolate.CreateTempDir(module, tempRoot, "cervomut-overlay-*")
	if err != nil {
		return "", "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(tmp) }
	rel, err := filepath.Rel(module, file)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		cleanup()
		return "", "", func() {}, errors.New("mutant file is outside module")
	}
	mutatedPath := filepath.Join(tmp, rel)
	if err := os.MkdirAll(filepath.Dir(mutatedPath), 0o755); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	if err := os.WriteFile(mutatedPath, data, 0o644); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	if err := ApplyReplacement(mutatedPath, startOffset, endOffset, original, mutated); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	overlayPath := filepath.Join(tmp, "overlay.json")
	overlay := struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{file: mutatedPath}}
	overlayData, err := json.MarshalIndent(overlay, "", "  ")
	if err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	if err := os.WriteFile(overlayPath, overlayData, 0o644); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	return module, overlayPath, cleanup, nil
}
