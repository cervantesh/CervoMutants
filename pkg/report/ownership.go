package report

import (
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/engine"
)

func hasOwnershipRoutes(mutants []engine.MutantResult) bool {
	for _, mutant := range mutants {
		if ownershipRouteConfigured(mutant.Mutant.Ownership) {
			return true
		}
	}
	return false
}

func ownershipRouteConfigured(route *engine.OwnershipRoute) bool {
	if route == nil {
		return false
	}
	return strings.TrimSpace(route.Owner) != "" ||
		strings.TrimSpace(route.Team) != "" ||
		strings.TrimSpace(route.Contact) != "" ||
		strings.TrimSpace(route.Rule) != ""
}

func ownershipRouteSummary(route *engine.OwnershipRoute) string {
	if !ownershipRouteConfigured(route) {
		return ""
	}
	parts := make([]string, 0, 4)
	if value := strings.TrimSpace(route.Owner); value != "" {
		parts = append(parts, "owner="+value)
	}
	if value := strings.TrimSpace(route.Team); value != "" {
		parts = append(parts, "team="+value)
	}
	if value := strings.TrimSpace(route.Contact); value != "" {
		parts = append(parts, "contact="+value)
	}
	if value := strings.TrimSpace(route.Rule); value != "" {
		parts = append(parts, "rule="+value)
	}
	return strings.Join(parts, " ")
}

func ownershipRouteOwner(route *engine.OwnershipRoute) string {
	if route == nil {
		return ""
	}
	return strings.TrimSpace(route.Owner)
}

func ownershipRouteTeam(route *engine.OwnershipRoute) string {
	if route == nil {
		return ""
	}
	return strings.TrimSpace(route.Team)
}

func ownershipRouteReviewOwner(route *engine.OwnershipRoute) string {
	for _, value := range []string{ownershipRouteOwner(route), ownershipRouteTeam(route)} {
		if value != "" {
			return value
		}
	}
	if route == nil {
		return ""
	}
	return strings.TrimSpace(route.Contact)
}

func ownershipRouteSearch(route *engine.OwnershipRoute) string {
	return ownershipRouteSummary(route)
}
