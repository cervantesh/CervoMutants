package engine

import (
	"path/filepath"
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/config"
)

func (e *Engine) ownershipRoute(pkgPath, filePath string) *OwnershipRoute {
	normalizedPkg := normalizeOwnershipPackage(pkgPath)
	normalizedFile := filepath.ToSlash(filePath)
	for _, rule := range e.cfg.Ownership.Rules {
		if ownershipRuleMatches(rule, normalizedPkg, normalizedFile) {
			return &OwnershipRoute{
				Owner:   strings.TrimSpace(rule.Owner),
				Team:    strings.TrimSpace(rule.Team),
				Contact: strings.TrimSpace(rule.Contact),
				Rule:    strings.TrimSpace(rule.Name),
			}
		}
	}
	if target := e.cfg.Ownership.Default; ownershipTargetConfigured(target) {
		return &OwnershipRoute{
			Owner:   strings.TrimSpace(target.Owner),
			Team:    strings.TrimSpace(target.Team),
			Contact: strings.TrimSpace(target.Contact),
			Rule:    "default",
		}
	}
	return nil
}

func ownershipRuleMatches(rule config.OwnershipRule, pkgPath, filePath string) bool {
	if selector := strings.TrimSpace(rule.Package); selector != "" && !ownershipPackageMatches(selector, pkgPath) {
		return false
	}
	if selector := strings.TrimSpace(rule.File); selector != "" && !suppressionFileMatches(selector, filePath) {
		return false
	}
	return true
}

func ownershipPackageMatches(pattern, pkgPath string) bool {
	pattern = normalizeOwnershipPackage(pattern)
	if pattern == "" {
		return false
	}
	if pkgPath == pattern {
		return true
	}
	return globMatch(pattern, pkgPath)
}

func normalizeOwnershipPackage(value string) string {
	value = filepath.ToSlash(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "./")
	if value == "" {
		return "."
	}
	return value
}

func ownershipTargetConfigured(target config.OwnershipTarget) bool {
	return strings.TrimSpace(target.Owner) != "" || strings.TrimSpace(target.Team) != "" || strings.TrimSpace(target.Contact) != ""
}
