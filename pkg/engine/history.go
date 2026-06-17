package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type historyFile struct {
	SchemaVersion string                  `json:"schema_version"`
	UpdatedAt     string                  `json:"updated_at"`
	Mutants       map[string]historyEntry `json:"mutants"`
	Runs          []HistoryRun            `json:"runs,omitempty"`
}

type historyEntry struct {
	MutantID         string `json:"mutant_id"`
	Operator         string `json:"operator"`
	Status           Status `json:"status"`
	FirstSeen        string `json:"first_seen"`
	LastSeen         string `json:"last_seen"`
	SeenRuns         int    `json:"seen_runs"`
	SurvivedRuns     int    `json:"survived_runs"`
	KilledRuns       int    `json:"killed_runs"`
	NotCoveredRuns   int    `json:"not_covered_runs"`
	CompileErrorRuns int    `json:"compile_error_runs"`
	TimedOutRuns     int    `json:"timed_out_runs"`
}

func (e *Engine) applyHistory(results []MutantResult) HistoryStats {
	stats := HistoryStats{Enabled: e.cfg.History.Enabled, Path: e.cfg.History.Path, OperatorUsefulSurvivor: map[string]float64{}}
	if !e.cfg.History.Enabled {
		return stats
	}
	path := e.historyPath()
	stats.Path = path
	store := e.loadHistoryStore(path)
	if store.Mutants == nil {
		store.Mutants = map[string]historyEntry{}
	}
	stats.UpdatedAt = store.UpdatedAt
	stats.Runs = append([]HistoryRun{}, store.Runs...)
	stats.LoadedMutants = len(store.Mutants)
	now := time.Now().UTC().Format(time.RFC3339)
	operatorSeen := map[string]int{}
	operatorSurvived := map[string]int{}
	for i := range results {
		result := &results[i]
		operator := historyOperator(result.Mutant.Operator)
		entry := updateHistoryResult(result, store.Mutants[result.MutantID], now, &stats)
		store.Mutants[result.MutantID] = entry
		operatorSeen[operator]++
		if result.Status == StatusSurvived {
			operatorSurvived[operator]++
		}
	}
	for operator, seen := range operatorSeen {
		if seen > 0 {
			stats.OperatorUsefulSurvivor[operator] = float64(operatorSurvived[operator]) / float64(seen)
		}
	}
	for i := range results {
		results[i].OperatorYield = stats.OperatorUsefulSurvivor[results[i].Mutant.Operator]
	}
	stats.UpdatedMutants = len(results)
	store.UpdatedAt = now
	stats.UpdatedAt = now
	e.writeHistoryStore(path, store)
	return stats
}

func (e *Engine) recordHistoryRun(result *RunResult) {
	if result == nil || !e.cfg.History.Enabled {
		return
	}
	path := e.historyPath()
	store := e.loadHistoryStore(path)
	run := historyRunFromResult(*result)
	store.Runs = append(store.Runs, run)
	store.UpdatedAt = run.RunAt
	e.writeHistoryStore(path, store)
	result.History.Path = path
	result.History.UpdatedAt = store.UpdatedAt
	result.History.Runs = append([]HistoryRun{}, store.Runs...)
}

func historyOperator(operator string) string {
	if operator == "" {
		return "unknown"
	}
	return operator
}

func updateHistoryResult(result *MutantResult, previous historyEntry, now string, stats *HistoryStats) historyEntry {
	if previous.MutantID == "" {
		result.FirstSeen = now
		if result.Status == StatusSurvived {
			result.HistoryStatus = "new_survivor"
			stats.NewSurvivors++
		}
	} else {
		result.PreviousStatus = previous.Status
		result.FirstSeen = previous.FirstSeen
		result.SurvivorAgeRuns = previous.SurvivedRuns
		markExistingSurvivor(result, previous, stats)
	}
	if result.HistoryStatus == "" {
		result.HistoryStatus = "seen"
	}
	result.LastSeen = now
	entry := updateHistoryEntry(previous, *result, now)
	result.SurvivorAgeRuns = entry.SurvivedRuns
	return entry
}

func markExistingSurvivor(result *MutantResult, previous historyEntry, stats *HistoryStats) {
	if result.Status != StatusSurvived {
		return
	}
	result.HistoryStatus = "existing_survivor"
	if previous.SurvivedRuns > 0 {
		stats.LongStandingSurvivors++
		result.HistoryStatus = "long_standing_survivor"
	}
}

func updateHistoryEntry(entry historyEntry, result MutantResult, now string) historyEntry {
	entry.MutantID = result.MutantID
	entry.Operator = historyOperator(result.Mutant.Operator)
	entry.Status = result.Status
	if entry.FirstSeen == "" {
		entry.FirstSeen = result.FirstSeen
	}
	entry.LastSeen = now
	entry.SeenRuns++
	incrementHistoryStatus(&entry, result.Status)
	return entry
}

func incrementHistoryStatus(entry *historyEntry, status Status) {
	switch status {
	case StatusSurvived:
		entry.SurvivedRuns++
	case StatusKilled:
		entry.KilledRuns++
	case StatusNotCovered:
		entry.NotCoveredRuns++
	case StatusCompileError:
		entry.CompileErrorRuns++
	case StatusTimedOut:
		entry.TimedOutRuns++
	}
}

func (e *Engine) historyPath() string {
	path := e.cfg.History.Path
	if path == "" {
		return ".cervomut/history.json"
	}
	return path
}

func (e *Engine) loadHistoryStore(path string) historyFile {
	store := historyFile{SchemaVersion: "1", Mutants: map[string]historyEntry{}}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &store)
	}
	if store.Mutants == nil {
		store.Mutants = map[string]historyEntry{}
	}
	return store
}

func (e *Engine) writeHistoryStore(path string, store historyFile) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err == nil {
		if data, err := json.MarshalIndent(store, "", "  "); err == nil {
			_ = os.WriteFile(path, data, 0o644)
		}
	}
}

func historyRunFromResult(result RunResult) HistoryRun {
	runAt := time.Now().UTC().Format(time.RFC3339)
	if lastSeen := latestResultTimestamp(result.Mutants); lastSeen != "" {
		runAt = lastSeen
	}
	newCount, agingCount, longStandingCount := survivorAgeCounts(result.Mutants)
	operatorYield := map[string]float64{}
	for operator, value := range result.History.OperatorUsefulSurvivor {
		operatorYield[operator] = value
	}
	return HistoryRun{
		RunAt:                   runAt,
		RawScore:                result.Summary.Score,
		ActionableScore:         result.Summary.Actionable.ActionableScore,
		Survived:                result.Summary.Survived,
		TrueActionableSurvivors: result.Summary.Actionable.TrueActionableSurvivors,
		NewSurvivors:            result.Summary.NewSurvivors,
		LongStandingSurvivors:   result.Summary.LongStandingSurvivors,
		SurvivorAgeNew:          newCount,
		SurvivorAgeAging:        agingCount,
		SurvivorAgeLongStanding: longStandingCount,
		TimedOut:                result.Summary.TimedOut,
		NonProgressTimeouts:     result.Summary.NonProgressTimeouts,
		OperatorUsefulSurvivor:  operatorYield,
	}
}

func latestResultTimestamp(results []MutantResult) string {
	latest := ""
	for _, result := range results {
		timestamp := result.LastSeen
		if timestamp == "" {
			continue
		}
		if latest == "" || timestamp > latest {
			latest = timestamp
		}
	}
	return latest
}

func survivorAgeCounts(results []MutantResult) (newCount, agingCount, longStandingCount int) {
	for _, result := range results {
		if result.Status != StatusSurvived {
			continue
		}
		switch {
		case result.SurvivorAgeRuns <= 1:
			newCount++
		case result.SurvivorAgeRuns < 5:
			agingCount++
		default:
			longStandingCount++
		}
	}
	return newCount, agingCount, longStandingCount
}
