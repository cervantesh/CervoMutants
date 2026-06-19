# Structured Adoption Feedback Wave From Released GitHub Action Artifacts

Tracking issue: [#292](https://github.com/cervantesh/cervo-mutants/issues/292)

Date: 2026-06-19

This note converts the released `github-action@v0.4.2` broader-profile hosted
sample into structured adoption-feedback issues. The goal is narrower than
"prove broad external maintainer adoption." The goal is to stop leaving this
evidence only in maintainer-authored study notes and move it into the public
intake path used by [docs/feedback-intake.md](../feedback-intake.md) and
[docs/adoption-analytics.md](../adoption-analytics.md).

Source release-wave evidence:

- released wave note:
  [2026-06-19-external-github-action-wave-v0.4.2-profile-sample.md](2026-06-19-external-github-action-wave-v0.4.2-profile-sample.md)
- aggregate artifact:
  [2026-06-19-external-github-action-wave-v0.4.2-profile-sample-summary.json](2026-06-19-external-github-action-wave-v0.4.2-profile-sample-summary.json)
- hosted run:
  `27848829038`

## Adoption-Feedback Issues Opened

| Repo | Profile | Main friction captured | Issue |
| --- | --- | --- | --- |
| `spf13/pflag` | compact library | healthy run, but equivalent-risk boundary survivors still need explicit review-once framing | [#294](https://github.com/cervantesh/cervo-mutants/issues/294) |
| `tidwall/gjson` | compact library | healthy run with semantic collapse; reviewers still need help interpreting repeated boundary survivors and report-only governance rows | [#295](https://github.com/cervantesh/cervo-mutants/issues/295) |
| `prometheus/prometheus` `./model/labels` | medium service | healthy zero-survivor run can still look like "no result" without denominator-aware interpretation | [#296](https://github.com/cervantesh/cervo-mutants/issues/296) |
| `kubernetes/apimachinery` `./pkg/api/resource` | large multipackage scoped target | healthy large-repo first run still surfaces several equivalent-risk boundary survivors that need bounded review semantics | [#297](https://github.com/cervantesh/cervo-mutants/issues/297) |

These issues now preserve the repository profile, adoption stage, install path,
environment, workflow posture, blocker class, and artifact evidence in the
canonical intake form instead of only in a release-wave summary.

## Cross-Issue Patterns

### 1. Equivalent-risk boundary survivors recur even on healthy runs

This is the clearest repeated pattern in the new issue set:

- `#294` (`pflag`) had healthy denominator behavior plus one
  high-equivalent-risk boundary survivor
- `#295` (`gjson`) had healthy denominator behavior plus a repeated len-boundary
  cluster that collapsed from `3` raw survivors to `2` review units
- `#297` (`apimachinery`) had healthy denominator behavior plus `3`
  high-equivalent-risk boundary survivors in one hot file

The correct interpretation is not "the hosted lane is noisy." The correct
interpretation is "healthy review-bearing runs still need explicit review-once
semantics for equivalent-risk boundary clusters."

### 2. Healthy zero-survivor runs need better interpretation than a raw score

`#296` (`prometheus-labels`) is useful because it shows the opposite shape:

- denominator health stayed healthy
- the run produced `7` effective mutants and killed all `7`
- `not_covered=3`
- there were no actionable survivors or recommendations

That is a valid first useful report, not a failure. It is healthy
rollout evidence with no immediate survivor review work.

### 3. The released hosted path itself was not the blocker

Across all four issues:

- no install failure
- no missing artifacts
- no compile errors
- no timeouts

The remaining friction was interpretation and review semantics, not basic
hosted-path viability.

## Promotion Decision

The repeated equivalent-risk boundary-survivor pattern crossed the promotion
threshold and was added to the
[follow-up ledger](follow-up-ledger.md).

The healthy zero-survivor interpretation case did **not** get its own ledger
entry from this wave alone. It matches earlier released-wave guidance, but this
issue set only contributes one new structured adoption-feedback issue of that
shape.

## Maturity Effect

This work improves the evidence format, not the maturity score by itself.

It does **not** prove independent external maintainer adoption across multiple
teams. It does prove that released-surface rollout evidence is now preserved in
the same structured public issue path that the project expects for future real
adoption feedback.
