# Released GitHub Action Validation Wave: v0.4.1

Tracking issue: [#258](https://github.com/cervantesh/cervo-mutants/issues/258)

Date: 2026-06-19

This document records the next external adoption wave that intentionally used
the current public GitHub Action release instead of the current branch source
for the action implementation.

The goals were specific:

1. prove that the hosted adoption wave still bootstraps from the released
   `v0.4.1` tag
2. persist committed released-surface evidence for denominator health,
   actionable review units, recommendations, semantic grouping, ledgers, and
   governance suggestions
3. determine whether the current public release changed review yield or exposed
   new rollout friction relative to the prior released wave

Committed aggregate artifact:

- [2026-06-19-external-github-action-wave-v0.4.1-summary.json](2026-06-19-external-github-action-wave-v0.4.1-summary.json)

## Inputs

Workflow and manifest under test:

- [.github/workflows/external-action-wave.yml](../../.github/workflows/external-action-wave.yml)
- [external-github-action-wave-v0.4.1-candidates.json](external-github-action-wave-v0.4.1-candidates.json)

Verification run:

- successful run: `27841757084`
- successful run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27841757084`
- branch carrying the manifest/default change: `codex/258-v041-adoption-wave`
- released surface under test: `github-action@v0.4.1`

## What Changed

Unlike the `v0.4.0` released-wave work, this follow-up did not need new
workflow plumbing. The hosted path had already been proven operational. The
actual branch change here was intentionally small:

1. add a `v0.4.1` external-wave manifest
2. point the workflow-dispatch default at that manifest so the next public wave
   targets the current release by default

That makes this run a cleaner product signal than the earlier release wave:
the evidence reflects the public release surface, not a release plus workflow
repair session.

## Released Surface Evidence

The successful run did not execute branch HEAD as the action under test. The
matrix jobs checked out the released tag for `cervomut-source`, and the
persisted artifacts carried the released identity:

- `action_ref: "v0.4.1"`
- `install_path: "github-action@v0.4.1"`

The committed aggregate summary now also preserves richer actionable-yield
fields than the earlier `v0.4.0` committed summary, including:

- `recommendation_review_units`
- `collapsed_recommendation_duplicates`
- `governance_suggestions_by_status`

That matters because `#258` is not only about "did the action run"; it is also
about preserving the released-surface review yield in a form maintainers can
audit later.

## Repositories

| Repo | Target | Profile | Policy | Budget | Max mutants |
| --- | --- | --- | --- | --- | ---: |
| `spf13/pflag` | `./...` | small library | `ci-balanced` | `5m` | `10` |
| `sirupsen/logrus` | `./...` | medium library | `ci-balanced` | `5m` | `10` |
| `tidwall/gjson` | `./...` | validation library | `ci-balanced` | `5m` | `10` |

All jobs ran on GitHub-hosted `ubuntu-latest` runners. The workflow branch only
supplied the updated manifest default; the mutation action itself was
bootstrapped from the released `v0.4.1` tag.

## Result

Aggregate result from run `27841757084`:

- selected repos: `3`
- reports captured: `3`
- missing reports: `0`
- generated mutants: `30`
- effective mutants: `24`
- killed: `18`
- survived: `6`
- not covered: `6`
- timed out: `0`
- compile errors: `0`
- actionable review units: `5`
- semantic group review units: `2`
- recommendation entries: `6`
- recommendation review units: `5`
- collapsed recommendation duplicates: `1`
- ledger entries: `3`
- governance suggestions: `8`
- repos with denominator warnings: `1`

Per-repo result:

| Repo | Job | Report | Generated | Effective | Killed | Survived | Not covered | Score | Actionable | Review units | Recommendations | Rec review units | Ledger | Denominator health |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `gjson-root` | `success` | `full` | `10` | `10` | `7` | `3` | `0` | `70` | `77.77777777777779` | `2` | `3` | `2` | `2` | `healthy` |
| `logrus-root` | `success` | `full` | `10` | `4` | `4` | `0` | `6` | `100` | `100` | `0` | `0` | `0` | `0` | `warning: not_covered_exceeds_effective, high_score_poor_denominator_health` |
| `pflag-root` | `success` | `full` | `10` | `10` | `7` | `3` | `0` | `70` | `70` | `3` | `3` | `3` | `1` | `healthy` |

## Findings

1. The current public GitHub Action release is operationally stable on the same
   bounded public sample used for the earlier released wave. The `v0.4.1` run
   completed successfully across all three repositories with no missing
   reports.
2. The top-line review yield is stable relative to the prior `v0.4.0`
   released-wave evidence. This wave produced the same bounded aggregate signal:
   `6` survivors, `5` actionable review units, `2` semantic-group review
   units, and `3` ledger entries.
3. The remaining weak repo is still weak for denominator-health reasons, not
   because semantic triage or recommendations collapsed. `logrus-root` again
   shows `not_covered` pressure and denominator warnings while correctly
   emitting zero actionable review units and zero recommendation entries.
4. The committed evidence is stronger than before because it now preserves
   recommendation-review-unit counts and status-split governance suggestions in
   the released-surface summary itself. That makes later rollout guidance less
   dependent on ad hoc artifact inspection.
5. This wave did not expose a new engine or release-path defect. The next
   justified follow-up is documentation and rollout guidance, especially around
   how to interpret denominator-poor runs and bounded hosted defaults. That is
   the scope of [#256](https://github.com/cervantesh/cervo-mutants/issues/256).

## Threats To Validity

1. This validates the current public hosted action path more strongly than it
   validates large or long-running mutation campaigns.
2. The wave is intentionally bounded to `10` mutants per repository with a `5m`
   budget, so it remains representative of CI-oriented adoption rather than
   deep exhaustive analysis.
3. All runs used Linux GitHub-hosted runners. This does not replace the broader
   compatibility matrix.
4. The sample remains narrow and branch/boundary heavy. It is enough to support
   practical rollout guidance, but not broad claims about every operator family
   or heuristic class.

## Conclusion

Issue `#258` now has the evidence it needed:

- the hosted adoption wave runs cleanly against the current public release
  `github-action@v0.4.1`
- the committed released-surface summary preserves actionable-yield fields that
  humans can audit later
- the review yield remains real and stable on the bounded public sample
- the next useful step is not more release-path validation, but tighter hosted
  rollout guidance grounded in these released-surface findings
