# 2026-06-19 Field Calibration From Adoption Waves

Tracking issue: [#211](https://github.com/cervantesh/cervo-mutants/issues/211)

This note evaluates semantic triage, actionable scoring, and recommendation
surfaces against the two real external adoption-wave artifacts currently
available in this repository.

The goal is not to pretend the calibration question is solved. The goal is to
separate what the current field evidence actually proves from what still
requires a better external sample.

## Evidence Base

Committed evidence used here:

- [2026-06-17-external-validation-wave.md](2026-06-17-external-validation-wave.md)
- [2026-06-17-external-validation-wave-summary.json](2026-06-17-external-validation-wave-summary.json)
- [2026-06-19-external-github-action-wave-2.md](2026-06-19-external-github-action-wave-2.md)
- [2026-06-19-external-github-action-wave-2-summary.json](2026-06-19-external-github-action-wave-2-summary.json)

Additional direct artifact inspection:

- hosted verification run `27831912391`
- per-repo artifacts downloaded from `external-wave-cobra-doc`,
  `external-wave-logrus-root`, and `external-wave-grpc-status`

## Wave Comparison

| Wave | Execution path | Repos | Mutants | Killed | Survived | Not covered | Main signal shape |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| 2026-06-17 external validation wave | local bounded mutation campaign | `5` | `50` | `37` | `13` | `0` | meaningful survivor variation with healthy denominators |
| 2026-06-19 hosted GitHub Action wave | first-party GitHub Action on GitHub-hosted runners | `3` | `10` generated, `0` effective on two repos and `0` total on one | `0` | `0` | `10` | operational success but calibration-poor survivor sample |

The important contrast is not simply "one wave scored higher". The important
contrast is that the first wave produced reviewable survivor variation, while
the hosted wave mostly produced denominator-poor rows.

## Artifact-Level Observations

Hosted wave 2 per-repo artifact inspection showed:

- `cobra-doc`
  - `total=5`
  - `not_covered=5`
  - `actionable_score=0`
  - `recommendation_count=0`
  - `semantic_group_count=0`
  - `semantic-triage-ledger.json` exists but `entries=[]`
- `logrus-root`
  - `total=5`
  - `not_covered=5`
  - `actionable_score=0`
  - `recommendation_count=0`
  - `semantic_group_count=0`
  - `semantic-triage-ledger.json` exists but `entries=[]`
- `grpc-status`
  - `total=0`
  - `actionable_score=0`
  - `recommendation_count=0`
  - `semantic_group_count=0`
  - `semantic-triage-ledger.json` exists but `entries=[]`

The hosted artifacts still generated the expected output surfaces:

- `mutation-report.json`
- `github-summary.md`
- `semantic-triage-ledger.json`
- `governance-review.md`
- `test-recommendations.md`

But those surfaces were empty or no-op because the run yielded no actionable
review units.

## Findings

### 1. Hosted action execution is now proven, but current hosted defaults are too weak for triage calibration

The hosted wave succeeded operationally: checkout, mutation run, artifact
upload, and summary aggregation all worked on GitHub-hosted runners.

That does **not** mean the hosted wave is already a good calibration source for
semantic triage or recommendation quality. With `policy=ci-fast`,
`max_mutants=5`, and the current target set:

- two repos produced only `not_covered` mutants
- one repo produced zero generated mutants
- all three repos ended with `actionable_score=0`

That means the next hosted-wave calibration focus should be input yield, not
fine-grained ranking-weight tuning.

### 2. The field evidence does not support saying recommendation quality is good or bad yet

The hosted wave produced:

- no actionable survivors
- no recommendation entries
- no semantic groups
- empty semantic triage ledgers

That is better than producing noisy false-positive recommendations, but it is
not enough evidence to claim the recommendation engine is already calibrated
against public-repo field data.

The correct conclusion is narrower:

- the recommendation and governance surfaces degrade gracefully when there are
  no actionable units
- field calibration of recommendation usefulness is still underpowered

### 3. Actionable scoring is behaving conservatively in denominator-poor cases

All hosted-wave repos landed at `actionable_score=0`.

Given the artifact shapes, that behavior is defensible:

- `cobra-doc` and `logrus-root` had no effective mutants
- `grpc-status` had no generated mutants at all

So the current field evidence does **not** indicate that actionable scoring is
ranking bad survivors ahead of good ones in hosted external waves. It indicates
that the current hosted wave does not yet produce enough reviewable units to
stress that ordering question.

### 4. External public repos can still produce meaningful survivor signal under bounded settings

The first external validation wave is the counterexample that keeps this from
turning into a vague complaint about public repos:

- `50` mutants executed
- `13` survivors
- `0` not-covered rows
- score range from `40` to `100`

That proves the current product can produce useful external review signal.
The calibration gap is therefore about the hosted wave shape and artifact
persistence, not about external validation being inherently too noisy.

### 5. The current committed wave summaries still under-document triage-specific yield

The 2026-06-17 public-wave summary persists raw mutation outcomes clearly, but
it does not preserve triage-specific fields such as:

- actionable review unit counts
- semantic group counts
- recommendation counts
- empty-versus-non-empty governance suggestions

That weakens future field calibration because later reviews have to reconstruct
those questions from temporary artifact directories instead of committed
evidence.

## Conclusions

1. The next calibration step should target **effective-mutant yield in hosted
   adoption waves**, not heuristic weight changes first.
2. Recommendation and semantic-triage surfaces currently look safe in the
   negative sense: they did not fabricate noisy guidance when the wave produced
   no actionable units.
3. The current field sample is still too weak to claim that recommendation
   quality or semantic grouping usefulness is calibrated for hosted external
   runs.
4. Future adoption-wave summaries should preserve more triage-specific yield
   fields so this analysis does not depend on ephemeral downloaded artifacts.

## Follow-Up Issues

- [#224](https://github.com/cervantesh/cervo-mutants/issues/224) Tune hosted
  external-action-wave defaults for effective-mutant yield
- [#223](https://github.com/cervantesh/cervo-mutants/issues/223) Persist
  triage and recommendation yield in adoption-wave summaries

## Practical Recommendation

For the next hosted external calibration pass:

1. keep the GitHub-hosted action path
2. widen the sample enough to produce real effective mutants
3. persist actionable-yield fields in the committed summary itself
4. only then revisit ranking-weight changes based on that stronger field sample

That sequence is more defensible than tuning semantic heuristics on top of a
hosted wave that currently collapses to zero actionable units.
