# 2026-06-19 Hosted Wave Triage-Yield Persistence

Tracking issue: [#223](https://github.com/cervantesh/cervo-mutants/issues/223)

Date: 2026-06-19

This note records the follow-up that made hosted adoption-wave summaries
persist triage-specific yield directly in the aggregate artifact instead of
forcing later reviews to download temporary per-repo artifacts.

The workflow verification run for this issue reused the hosted default manifest
from [#224](https://github.com/cervantesh/cervo-mutants/issues/224), so the
summary artifact still reports `tracking_issue: "#224"` from the manifest
itself. The change under test here is the **summary shape**, not the hosted
candidate set.

Committed aggregate artifact:

- [2026-06-19-external-github-action-wave-triage-yield-summary.json](2026-06-19-external-github-action-wave-triage-yield-summary.json)

## Inputs

Workflow under test:

- [.github/workflows/external-action-wave.yml](../../.github/workflows/external-action-wave.yml)

Exact verification run:

- run id: `27833492410`
- run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27833492410`
- branch: `codex/223-adoption-wave-triage-yield`
- workflow result: `success`

## What Changed

The hosted wave summary now persists a new additive `triage` block:

- per repository under each `repos[]` item
- aggregated again at the top level beside the existing raw totals

The new summary block captures:

- `actionable_review_units`
- `actionable_survivors`
- `equivalent_risk_survivors`
- `semantic_group_review_units`
- `collapsed_semantic_duplicates`
- `semantic_group_count`
- `recommendation_entries`
- `ledger_entries`
- `governance_quarantine_templates`
- `governance_suppression_templates`
- `governance_total_suggestions`

The aggregate block also records repo-level presence counts such as:

- `repos_with_actionable_review_units`
- `repos_with_semantic_groups`
- `repos_with_recommendations`
- `repos_with_ledger_entries`
- `repos_with_governance_suggestions`

## Why This Matters

Before this change, the raw mutation totals were preserved well, but later
field calibration had to inspect downloaded artifacts manually to answer
questions like:

- were any recommendation entries generated?
- was the semantic triage ledger empty?
- were governance suggestions present even when survivors were not actionable?

After this change, those questions are answerable from the committed summary
artifact itself.

## Observed Result

Aggregate raw result in verification run `27833492410`:

- generated: `30`
- effective: `4`
- killed: `4`
- survived: `0`
- not covered: `26`

Aggregate triage result now persisted in the same summary:

- actionable review units: `0`
- semantic groups formed: `0`
- recommendation entries: `0`
- ledger entries: `0`
- governance total suggestions: `7`
- repos with governance suggestions: `3`

That is the important proof point for `#223`: the summary can now preserve
useful negative and positive triage signals at the same time. In this hosted
wave:

- recommendation and semantic-group yield stayed empty
- the ledger stayed empty
- governance suggestions were still present and countable

Previously that last distinction was not visible from the committed aggregate
summary alone.

## Conclusion

`#223` is not about improving the hosted wave's mutation quality directly.
It is about making future adoption-wave evidence self-describing enough for
calibration review.

This pass satisfies that narrower goal:

- the summary shape stays additive
- raw denominator totals remain intact
- triage, recommendation, ledger, and governance yield are now committed in the
  same aggregate artifact

That means future calibration notes no longer need ephemeral artifact downloads
just to answer baseline triage-yield questions.
