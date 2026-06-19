# Governance Suggestion Status Counts For Adoption-Wave Summaries

Tracking issue: [#243](https://github.com/cervantesh/cervo-mutants/issues/243)

Date: 2026-06-19

This note records the additive summary change that separates raw governance
suggestion count from the mutant statuses that produced those suggestions in
hosted adoption-wave artifacts.

## Why This Exists

The released `v0.4.0` hosted wave showed that governance suggestions are not
the same thing as survivor review workload.

The concrete mismatch was `logrus-root`:

- `survived=0`
- `actionable_review_units=0`
- `governance_total_suggestions=2`

Those two governance entries did not represent hidden survivor work. They came
from `not_covered` mutants that still produced audit-oriented governance
templates.

That means the existing raw governance total is still useful, but it is too
easy to over-read without knowing which mutant states produced it.

## Change

The hosted `external-action-wave.yml` summary now preserves two related
governance views:

- `governance_total_suggestions`
  - the unchanged raw total across suppression and quarantine templates
- `governance_suggestions_by_status`
  - an additive count map grouped by mutant status such as `survived`,
    `killed`, or `not_covered`

These fields are added:

- per repository under each `repos[].triage` block
- aggregated again under the top-level `triage` block

The existing raw `governance_total_suggestions` field remains unchanged.

## Derivation Rule

The new status map is derived directly from the same governance inputs already
emitted by the product:

- `governance-review.json`
- `suppression_templates[]`
- `quarantine_templates[]`
- each template's existing `status` field

This keeps the new summary grounded in the current governance artifact shape
instead of introducing a second interpretation layer.

## Verification Run

Hosted verification run:

- run id: `27837792047`
- run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27837792047`
- branch: `codex/243-governance-status-counts`
- manifest:
  `docs/evaluations/external-github-action-wave-v0.4.0-candidates.json`
- workflow result: `success`

Observed aggregate triage values in the generated `wave-summary.json`:

- `governance_total_suggestions=8`
- `governance_suggestions_by_status.survived=4`
- `governance_suggestions_by_status.killed=2`
- `governance_suggestions_by_status.not_covered=2`

Per-repo result:

| Repo | Governance total | Status counts | Why it matters |
| --- | ---: | --- | --- |
| `gjson-root` | `4` | `killed=1, survived=3` | governance mostly tracks real survivor follow-up here, but not exclusively |
| `logrus-root` | `2` | `not_covered=2` | proves governance can stay non-zero even when survivor workload is zero |
| `pflag-root` | `2` | `killed=1, survived=1` | shows mixed governance signal inside a healthy survivor-producing repo |

## Interpretation

These counts should now be read together:

- `governance_total_suggestions` answers:
  - how many governance templates were emitted in total
- `governance_suggestions_by_status` answers:
  - which mutant states produced those templates

That makes the adoption-wave summary materially easier to read:

- non-zero governance totals no longer imply hidden survivor work
- `not_covered` and `killed` governance hints remain visible as audit signal
- survivor review workload can stay tied to `actionable_review_units` instead of
  being inferred from governance totals alone
