package coverageprofile

import (
	"path/filepath"
	"strconv"
	"strings"
)

func DataSignal(data, rel, base string, mutantLine int) (lineCovered bool, fileCovered bool) {
	parseable := false
	for _, line := range strings.Split(data, "\n") {
		file, startLine, endLine, count, ok := ParseLine(line)
		if !ok {
			continue
		}
		parseable = true
		if count <= 0 || !FileMatches(file, rel, base) {
			continue
		}
		fileCovered = true
		if mutantLine >= startLine && mutantLine <= endLine {
			lineCovered = true
		}
	}
	if !parseable {
		fileCovered = FallbackMentions(data, rel, base)
		lineCovered = fileCovered
	}
	return lineCovered, fileCovered
}

func ParseLine(line string) (string, int, int, int, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "mode:") {
		return "", 0, 0, 0, false
	}
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return "", 0, 0, 0, false
	}
	colon := strings.LastIndex(fields[0], ":")
	if colon < 0 {
		return "", 0, 0, 0, false
	}
	file := filepath.ToSlash(fields[0][:colon])
	span := fields[0][colon+1:]
	comma := strings.Index(span, ",")
	if comma < 0 {
		return "", 0, 0, 0, false
	}
	startLine, ok := parseLineNumber(span[:comma])
	if !ok {
		return "", 0, 0, 0, false
	}
	endLine, ok := parseLineNumber(span[comma+1:])
	if !ok {
		return "", 0, 0, 0, false
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return "", 0, 0, 0, false
	}
	return file, startLine, endLine, count, true
}

func FileMatches(profileFile, rel, base string) bool {
	profileFile = filepath.ToSlash(profileFile)
	return profileFile == rel || strings.HasSuffix(profileFile, "/"+rel) || filepath.Base(profileFile) == base
}

func FallbackMentions(data, rel, base string) bool {
	for _, line := range strings.Split(data, "\n") {
		if strings.Contains(line, rel+":") || strings.Contains(line, base+":") {
			return true
		}
	}
	return false
}

func parseLineNumber(value string) (int, bool) {
	dot := strings.Index(value, ".")
	if dot < 0 {
		return 0, false
	}
	line, err := strconv.Atoi(value[:dot])
	return line, err == nil
}
