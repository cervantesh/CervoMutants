package engine

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cervantesh/cervo-mutants/pkg/quarantine"
)

func (s *runSession) loadQuarantine() (map[string]bool, int, error) {
	active := map[string]bool{}
	if !s.engine.cfg.Quarantine.Enabled {
		return active, 0, nil
	}
	data, err := os.ReadFile(s.engine.cfg.Quarantine.Path)
	if os.IsNotExist(err) {
		return active, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	var entries []quarantine.Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, 0, err
	}
	policy := quarantine.Policy{
		RequireReason: s.engine.cfg.Quarantine.RequireReason,
		RequireOwner:  s.engine.cfg.Quarantine.RequireOwner,
		RequireIssue:  s.engine.cfg.Quarantine.RequireIssue,
		FailOnExpired: s.engine.cfg.Quarantine.FailOnExpired,
		MaxRenewals:   s.engine.cfg.Quarantine.MaxRenewals,
	}
	now := time.Now()
	if err := quarantine.Validate(entries, policy, now); err != nil {
		return nil, 0, err
	}
	expired := 0
	for _, entry := range entries {
		if entry.ExpiresAt.After(now) {
			active[entry.MutantID] = true
		} else {
			expired++
		}
	}
	return active, expired, nil
}
