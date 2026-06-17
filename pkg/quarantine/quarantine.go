package quarantine

import (
	"errors"
	"strings"
	"time"
)

type Entry struct {
	MutantID  string    `json:"mutant_id"`
	Reason    string    `json:"reason"`
	Owner     string    `json:"owner"`
	Issue     string    `json:"issue"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Renewals  int       `json:"renewals"`
}

type Policy struct {
	RequireReason bool `json:"require_reason"`
	RequireOwner  bool `json:"require_owner"`
	RequireIssue  bool `json:"require_issue"`
	FailOnExpired bool `json:"fail_on_expired"`
	MaxRenewals   int  `json:"max_renewals"`
}

func Validate(entries []Entry, policy Policy, now time.Time) error {
	for _, entry := range entries {
		if err := validateEntry(entry, policy, now); err != nil {
			return err
		}
	}
	return nil
}

func validateEntry(entry Entry, policy Policy, now time.Time) error {
	if strings.TrimSpace(entry.MutantID) == "" {
		return errors.New("quarantine entry requires mutant_id")
	}
	required := []struct {
		enabled bool
		value   string
		message string
	}{
		{policy.RequireReason, entry.Reason, "quarantine entry requires reason"},
		{policy.RequireOwner, entry.Owner, "quarantine entry requires owner"},
		{policy.RequireIssue, entry.Issue, "quarantine entry requires issue"},
	}
	for _, field := range required {
		if field.enabled && strings.TrimSpace(field.value) == "" {
			return errors.New(field.message)
		}
	}
	if err := validateExpiry(entry, policy, now); err != nil {
		return err
	}
	if policy.MaxRenewals > 0 && entry.Renewals > policy.MaxRenewals {
		return errors.New("quarantine entry exceeded max renewals")
	}
	return nil
}

func validateExpiry(entry Entry, policy Policy, now time.Time) error {
	if entry.ExpiresAt.IsZero() {
		return errors.New("quarantine entry requires expires_at")
	}
	if policy.FailOnExpired && !entry.ExpiresAt.After(now) {
		return errors.New("quarantine entry expired")
	}
	return nil
}
