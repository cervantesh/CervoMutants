package gotestenv

import (
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/config"
)

type Plan struct {
	Env        []string
	Applied    bool
	GoFlags    string
	GOMAXPROCS string
}

func EffectiveCommandEnv(goos, isolation string, workers int, command, baseEnv []string) Plan {
	if goos != "windows" || !IsGoTestCommand(command) {
		return Plan{Env: append([]string{}, baseEnv...)}
	}
	if isolation != config.IsolationTempWorkdir && workers <= 2 {
		return Plan{Env: append([]string{}, baseEnv...)}
	}
	values := map[string]string{}
	order := make([]string, 0, len(baseEnv))
	for _, entry := range baseEnv {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if _, exists := values[name]; !exists {
			order = append(order, name)
		}
		values[name] = value
	}
	goFlags := NormalizeGoFlags(values["GOFLAGS"])
	values["GOFLAGS"] = goFlags
	maxProcs := "1"
	if workers > 1 {
		maxProcs = "2"
	}
	values["GOMAXPROCS"] = maxProcs
	if !containsString(order, "GOFLAGS") {
		order = append(order, "GOFLAGS")
	}
	if !containsString(order, "GOMAXPROCS") {
		order = append(order, "GOMAXPROCS")
	}
	env := make([]string, 0, len(order))
	for _, name := range order {
		env = append(env, name+"="+values[name])
	}
	return Plan{
		Env:        env,
		Applied:    true,
		GoFlags:    goFlags,
		GOMAXPROCS: maxProcs,
	}
}

func NormalizeGoFlags(current string) string {
	fields := strings.Fields(current)
	filtered := make([]string, 0, len(fields)+1)
	skipValue := false
	for _, field := range fields {
		if skipValue {
			skipValue = false
			continue
		}
		if field == "-p" {
			skipValue = true
			continue
		}
		if strings.HasPrefix(field, "-p=") {
			continue
		}
		filtered = append(filtered, field)
	}
	filtered = append(filtered, "-p=1")
	return strings.Join(filtered, " ")
}

func IsGoTestCommand(command []string) bool {
	return len(command) >= 2 && command[0] == "go" && command[1] == "test"
}

func PackageScopedCommand(command []string, pkg string) []string {
	next := append([]string{}, command[:2]...)
	replacedPackage := false
	for i := 2; i < len(command); i++ {
		arg := command[i]
		if isGoTestFlagWithSeparateValue(arg) && i+1 < len(command) {
			next = append(next, arg, command[i+1])
			i++
			continue
		}
		if strings.HasPrefix(arg, "-") {
			next = append(next, arg)
			continue
		}
		if !replacedPackage {
			next = append(next, pkg)
			replacedPackage = true
		}
	}
	if !replacedPackage {
		next = append(next, pkg)
	}
	return next
}

func WithCoverProfile(command []string, profile string) []string {
	if !IsGoTestCommand(command) {
		return command
	}
	next := append([]string{}, command...)
	for _, arg := range next[2:] {
		if strings.HasPrefix(arg, "-coverprofile") {
			return next
		}
	}
	return append(next[:2], append([]string{"-coverprofile", profile}, next[2:]...)...)
}

func WithOverlayFlag(command []string, overlayPath string) []string {
	next := append([]string{}, command...)
	if IsGoTestCommand(next) {
		return append(append([]string{}, next[:2]...), append([]string{"-overlay", overlayPath}, next[2:]...)...)
	}
	return next
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func isGoTestFlagWithSeparateValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	switch arg {
	case "-run", "-bench", "-count", "-timeout", "-coverprofile", "-covermode", "-coverpkg", "-tags", "-cpu", "-parallel", "-shuffle":
		return true
	default:
		return false
	}
}
