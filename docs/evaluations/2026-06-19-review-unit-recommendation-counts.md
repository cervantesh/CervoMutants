# Review-Unit Recommendation Counts For Adoption-Wave Summaries

Tracking issue: [#242](https://github.com/cervantesh/cervo-mutants/issues/242)

Date: 2026-06-19

This note records the additive summary change that separates raw recommendation
count from review-unit recommendation workload in hosted adoption-wave
artifacts.

## Why This Exists

The released `v0.4.0` hosted wave produced a small but important mismatch:

- raw `recommendation_entries=6`
- real review-unit recommendation workload=`5`
- collapsed recommendation duplicates=`1`

The concrete source was `gjson-root`, where three surviving mutants carried
recommendations but only two independent review units remained after semantic
group collapse.

That means raw recommendation count is still useful, but it is not the best
single proxy for how many review actions a human actually has to process.

## Change

The hosted `external-action-wave.yml` summary now preserves three related
fields:

- `recommendation_entries`
  - raw count of mutants with a non-empty `test_recommendation`
- `recommendation_review_units`
  - count of independent review units after collapsing recommendation-bearing
    semantic-group duplicates
- `collapsed_recommendation_duplicates`
  - the difference between those two counts

These fields are added:

- per repository under each `repos[].triage` block
- aggregated again under the top-level `triage` block

The existing `recommendation_entries` field remains unchanged.

## Derivation Rule

The new review-unit count is derived from real summary inputs already present in
`mutation-report.json`:

- mutants without `mutant.semantic_group` count as one review unit each
- mutants with the same `mutant.semantic_group` collapse to one review unit

That keeps the new count grounded in the same semantic-group keys already used
elsewhere in the product, instead of inventing a second grouping heuristic.

## Verification Run

Hosted verification run:

- run id: `27837358116`
- run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27837358116`
- branch: `codex/242-review-unit-recommendations`
- manifest:
  `docs/evaluations/external-github-action-wave-v0.4.0-candidates.json`
- workflow result: `success`

Observed aggregate triage values in the generated `wave-summary.json`:

- `recommendation_entries=6`
- `recommendation_review_units=5`
- `collapsed_recommendation_duplicates=1`

Per-repo result:

| Repo | Recommendation entries | Recommendation review units | Collapsed duplicates |
| --- | ---: | ---: | ---: |
| `gjson-root` | `3` | `2` | `1` |
| `logrus-root` | `0` | `0` | `0` |
| `pflag-root` | `3` | `3` | `0` |

## Interpretation

These counts should now be read together:

- `recommendation_entries` answers:
  - how many mutants emitted recommendation payloads
- `recommendation_review_units` answers:
  - how many independent review actions remain after semantic collapse
- `collapsed_recommendation_duplicates` answers:
  - how much raw recommendation count overstates human review workload

That is a narrower and more honest metric surface for adoption-wave analysis
than treating raw recommendation count as if it were already review-unit aware.
