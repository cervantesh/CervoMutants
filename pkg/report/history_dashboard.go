package report

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/cervantesh/cervo-mutants/pkg/engine"
)

type HistoryDashboard struct {
	SchemaVersion string                     `json:"schema_version"`
	RunCount      int                        `json:"run_count"`
	Latest        *engine.HistoryRun         `json:"latest,omitempty"`
	Previous      *engine.HistoryRun         `json:"previous,omitempty"`
	Delta         *HistoryDashboardDelta     `json:"delta,omitempty"`
	Runs          []engine.HistoryRun        `json:"runs,omitempty"`
	TopOperators  []HistoryDashboardOperator `json:"top_operator_yield,omitempty"`
}

type HistoryDashboardDelta struct {
	RawScore                float64 `json:"raw_score"`
	ActionableScore         float64 `json:"actionable_score"`
	Survived                int     `json:"survived"`
	TrueActionableSurvivors int     `json:"true_actionable_survivors"`
	TimedOut                int     `json:"timed_out"`
	NonProgressTimeouts     int     `json:"non_progress_timeouts"`
	SurvivorAgeNew          int     `json:"survivor_age_new"`
	SurvivorAgeAging        int     `json:"survivor_age_aging"`
	SurvivorAgeLongStanding int     `json:"survivor_age_long_standing"`
}

type HistoryDashboardOperator struct {
	Operator string  `json:"operator"`
	Yield    float64 `json:"yield"`
}

func HistoryDashboardJSON(result engine.RunResult) ([]byte, error) {
	return json.MarshalIndent(buildHistoryDashboard(result), "", "  ")
}

func HistorySummary(result engine.RunResult) string {
	dashboard := buildHistoryDashboard(result)
	if dashboard.Latest == nil {
		return "No historical dashboard data recorded yet.\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Historical runs: %d\n", dashboard.RunCount)
	fmt.Fprintf(&b, "Latest run: %s\n", fallbackText(dashboard.Latest.RunAt, "unknown"))
	fmt.Fprintf(&b, "Latest raw score: %.2f%%\n", dashboard.Latest.RawScore)
	fmt.Fprintf(&b, "Latest actionable score: %.2f%%\n", dashboard.Latest.ActionableScore)
	fmt.Fprintf(&b, "Latest survivors: %d\n", dashboard.Latest.Survived)
	fmt.Fprintf(&b, "Latest true actionable survivors: %d\n", dashboard.Latest.TrueActionableSurvivors)
	fmt.Fprintf(&b, "Survivor aging: new=%d aging=%d long-standing=%d\n", dashboard.Latest.SurvivorAgeNew, dashboard.Latest.SurvivorAgeAging, dashboard.Latest.SurvivorAgeLongStanding)
	fmt.Fprintf(&b, "Timeout trend: timed_out=%d non_progress=%d\n", dashboard.Latest.TimedOut, dashboard.Latest.NonProgressTimeouts)
	if dashboard.Delta != nil {
		fmt.Fprintf(&b, "Delta vs previous: raw=%+.2f actionable=%+.2f survived=%+d true_actionable=%+d timed_out=%+d long_standing=%+d\n",
			dashboard.Delta.RawScore,
			dashboard.Delta.ActionableScore,
			dashboard.Delta.Survived,
			dashboard.Delta.TrueActionableSurvivors,
			dashboard.Delta.TimedOut,
			dashboard.Delta.SurvivorAgeLongStanding,
		)
	}
	if len(dashboard.TopOperators) > 0 {
		b.WriteString("Top operator yield:\n")
		for _, operator := range dashboard.TopOperators {
			fmt.Fprintf(&b, "- %s: %.2f\n", operator.Operator, operator.Yield)
		}
	}
	return b.String()
}

func HistoryDashboardHTML(result engine.RunResult) string {
	dashboard := buildHistoryDashboard(result)
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\">")
	b.WriteString("<title>cervomut history dashboard</title>")
	b.WriteString(`<style>
body{margin:0;font-family:Segoe UI,Arial,sans-serif;background:#f3f6fb;color:#162033}
.page{max-width:1440px;margin:0 auto;padding:24px}
.hero{display:grid;gap:14px;margin-bottom:20px}
.hero h1{margin:0;font-size:30px}
.hero p{margin:0;max-width:960px;color:#44506a}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px;margin-bottom:18px}
.card,.panel{background:#fff;border:1px solid #d9e1f2;border-radius:14px;padding:16px;box-shadow:0 4px 18px rgba(22,32,51,.05)}
.card-label{display:block;font-size:12px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:#61708e}
.card-value{display:block;margin-top:8px;font-size:28px;font-weight:700}
.grid{display:grid;grid-template-columns:2fr 1fr;gap:16px}
.panel h2{margin:0 0 12px 0;font-size:18px}
table{width:100%;border-collapse:collapse}
th,td{padding:10px 8px;border-top:1px solid #e2e8f4;text-align:left;vertical-align:top}
thead th{border-top:0;font-size:12px;text-transform:uppercase;letter-spacing:.04em;color:#61708e}
.delta-positive{color:#176338;font-weight:700}
.delta-negative{color:#8d261d;font-weight:700}
.delta-neutral{color:#44506a;font-weight:700}
.empty{color:#61708e}
@media (max-width:1024px){.page{padding:16px}.grid{grid-template-columns:1fr}.panel{overflow:auto}}
</style></head><body><div class="page">`)
	b.WriteString(`<section class="hero"><div><h1>cervomut history dashboard</h1><p>Historical analytics view for raw score, actionable score, survivor aging, timeout movement, and operator yield across recorded runs. This dashboard is generated from the persisted run history emitted by the engine.</p></div></section>`)
	if dashboard.Latest == nil {
		b.WriteString(`<div class="panel"><p class="empty">No historical dashboard data recorded yet.</p></div></div></body></html>`)
		return b.String()
	}
	b.WriteString(`<section class="cards">`)
	writeHTMLCard(&b, "Historical runs", dashboard.RunCount)
	writeHTMLCard(&b, "Latest raw score", fmt.Sprintf("%.2f%%", dashboard.Latest.RawScore))
	writeHTMLCard(&b, "Latest actionable score", fmt.Sprintf("%.2f%%", dashboard.Latest.ActionableScore))
	writeHTMLCard(&b, "Current survivors", dashboard.Latest.Survived)
	writeHTMLCard(&b, "True actionable survivors", dashboard.Latest.TrueActionableSurvivors)
	writeHTMLCard(&b, "Long-standing survivors", dashboard.Latest.SurvivorAgeLongStanding)
	writeHTMLCard(&b, "Timed out", dashboard.Latest.TimedOut)
	writeHTMLCard(&b, "Non-progress timeouts", dashboard.Latest.NonProgressTimeouts)
	b.WriteString(`</section>`)
	b.WriteString(`<section class="grid"><div class="panel"><h2>Run trend</h2><table><thead><tr><th>Run</th><th>Raw score</th><th>Actionable score</th><th>Survivors</th><th>True actionable</th><th>Age new</th><th>Age 2-4</th><th>Age 5+</th><th>Timed out</th><th>Non-progress</th></tr></thead><tbody>`)
	for _, run := range dashboard.Runs {
		fmt.Fprintf(&b, `<tr><td>%s</td><td>%.2f%%</td><td>%.2f%%</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td><td>%d</td></tr>`,
			html.EscapeString(fallbackText(run.RunAt, "unknown")),
			run.RawScore,
			run.ActionableScore,
			run.Survived,
			run.TrueActionableSurvivors,
			run.SurvivorAgeNew,
			run.SurvivorAgeAging,
			run.SurvivorAgeLongStanding,
			run.TimedOut,
			run.NonProgressTimeouts,
		)
	}
	b.WriteString(`</tbody></table></div><div class="panel"><h2>Current deltas</h2>`)
	if dashboard.Delta == nil {
		b.WriteString(`<p class="empty">Only one historical run is available.</p>`)
	} else {
		writeHistoryDelta(&b, "Raw score", dashboard.Delta.RawScore, "%")
		writeHistoryDelta(&b, "Actionable score", dashboard.Delta.ActionableScore, "%")
		writeHistoryDelta(&b, "Survivors", float64(dashboard.Delta.Survived), "")
		writeHistoryDelta(&b, "True actionable survivors", float64(dashboard.Delta.TrueActionableSurvivors), "")
		writeHistoryDelta(&b, "Timed out", float64(dashboard.Delta.TimedOut), "")
		writeHistoryDelta(&b, "Long-standing survivors", float64(dashboard.Delta.SurvivorAgeLongStanding), "")
	}
	if len(dashboard.TopOperators) > 0 {
		b.WriteString(`<h2>Latest operator yield</h2><table><thead><tr><th>Operator</th><th>Yield</th></tr></thead><tbody>`)
		for _, operator := range dashboard.TopOperators {
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%.2f</td></tr>`, html.EscapeString(operator.Operator), operator.Yield)
		}
		b.WriteString(`</tbody></table>`)
	}
	b.WriteString(`</div></section></div></body></html>`)
	return b.String()
}

func buildHistoryDashboard(result engine.RunResult) HistoryDashboard {
	runs := historyRuns(result)
	dashboard := HistoryDashboard{
		SchemaVersion: "1",
		RunCount:      len(runs),
		Runs:          runs,
	}
	if len(runs) == 0 {
		return dashboard
	}
	latest := runs[len(runs)-1]
	dashboard.Latest = &latest
	if len(runs) > 1 {
		previous := runs[len(runs)-2]
		dashboard.Previous = &previous
		dashboard.Delta = &HistoryDashboardDelta{
			RawScore:                latest.RawScore - previous.RawScore,
			ActionableScore:         latest.ActionableScore - previous.ActionableScore,
			Survived:                latest.Survived - previous.Survived,
			TrueActionableSurvivors: latest.TrueActionableSurvivors - previous.TrueActionableSurvivors,
			TimedOut:                latest.TimedOut - previous.TimedOut,
			NonProgressTimeouts:     latest.NonProgressTimeouts - previous.NonProgressTimeouts,
			SurvivorAgeNew:          latest.SurvivorAgeNew - previous.SurvivorAgeNew,
			SurvivorAgeAging:        latest.SurvivorAgeAging - previous.SurvivorAgeAging,
			SurvivorAgeLongStanding: latest.SurvivorAgeLongStanding - previous.SurvivorAgeLongStanding,
		}
	}
	dashboard.TopOperators = topHistoryOperators(latest.OperatorUsefulSurvivor, 8)
	return dashboard
}

func historyRuns(result engine.RunResult) []engine.HistoryRun {
	if len(result.History.Runs) > 0 {
		runs := append([]engine.HistoryRun{}, result.History.Runs...)
		sort.SliceStable(runs, func(i, j int) bool {
			return runs[i].RunAt < runs[j].RunAt
		})
		return runs
	}
	if result.Summary.Total == 0 && result.Summary.GeneratedMutants == 0 {
		return nil
	}
	return []engine.HistoryRun{fallbackHistoryRun(result)}
}

func fallbackHistoryRun(result engine.RunResult) engine.HistoryRun {
	runAt := result.History.UpdatedAt
	if runAt == "" {
		for _, mutant := range result.Mutants {
			if mutant.LastSeen != "" && mutant.LastSeen > runAt {
				runAt = mutant.LastSeen
			}
		}
	}
	newCount, agingCount, longStandingCount := 0, 0, 0
	for _, mutant := range result.Mutants {
		if mutant.Status != engine.StatusSurvived {
			continue
		}
		switch {
		case mutant.SurvivorAgeRuns <= 1:
			newCount++
		case mutant.SurvivorAgeRuns < 5:
			agingCount++
		default:
			longStandingCount++
		}
	}
	operatorYield := map[string]float64{}
	for operator, value := range result.History.OperatorUsefulSurvivor {
		operatorYield[operator] = value
	}
	return engine.HistoryRun{
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

func topHistoryOperators(values map[string]float64, limit int) []HistoryDashboardOperator {
	if len(values) == 0 || limit <= 0 {
		return nil
	}
	operators := make([]HistoryDashboardOperator, 0, len(values))
	for operator, value := range values {
		operators = append(operators, HistoryDashboardOperator{Operator: operator, Yield: value})
	}
	sort.SliceStable(operators, func(i, j int) bool {
		if operators[i].Yield != operators[j].Yield {
			return operators[i].Yield > operators[j].Yield
		}
		return operators[i].Operator < operators[j].Operator
	})
	if len(operators) > limit {
		operators = operators[:limit]
	}
	return operators
}

func writeHistoryDelta(b *strings.Builder, label string, value float64, suffix string) {
	className := "delta-neutral"
	switch {
	case value > 0:
		className = "delta-positive"
	case value < 0:
		className = "delta-negative"
	}
	fmt.Fprintf(b, `<p><strong>%s:</strong> <span class="%s">%+.2f%s</span></p>`,
		html.EscapeString(label),
		className,
		value,
		html.EscapeString(suffix),
	)
}
