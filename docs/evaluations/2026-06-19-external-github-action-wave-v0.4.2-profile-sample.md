# Released GitHub Action Validation Wave: v0.4.2 Profile Sample

Tracking issue: [#288](https://github.com/cervantesh/cervo-mutants/issues/288)

Date: 2026-06-19

This document records the next released-surface GitHub Action wave after the
bounded `pflag/gjson/logrus` sample on `v0.4.2`. The goal here was not to
re-prove that the hosted path works. That was already established in
[#284](https://github.com/cervantesh/cervo-mutants/issues/284). The goal was
narrower:

1. keep the released `github-action@v0.4.2` surface under test
2. refresh the broader-profile released sample that previously centered on
   `v0.4.1`
3. use the calibrated broader-profile targets proven by
   [#265](https://github.com/cervantesh/cervo-mutants/issues/265) and
   [#269](https://github.com/cervantesh/cervo-mutants/issues/269)
4. measure whether the current public release now produces real mixed-profile
   review signal instead of mostly denominator-poor evidence

Committed aggregate artifact:

- [2026-06-19-external-github-action-wave-v0.4.2-profile-sample-summary.json](2026-06-19-external-github-action-wave-v0.4.2-profile-sample-summary.json)

## Inputs

Workflow and manifest under test:

- [.github/workflows/external-action-wave.yml](../../.github/workflows/external-action-wave.yml)
- [external-github-action-wave-v0.4.2-profile-sample.json](external-github-action-wave-v0.4.2-profile-sample.json)

Verification run:

- successful run: `27848829038`
- successful run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27848829038`
- branch carrying the manifest:
  `codex/288-v042-broader-profile-wave`
- released surface under test: `github-action@v0.4.2`

## Candidate Mix

This wave intentionally mixed smaller control repos with the broader-profile
targets that had already been recalibrated and toolchain-revalidated:

| Repo | Target | Profile | Special note |
| --- | --- | --- | --- |
| `spf13/pflag` | `./...` | `small-library-control` | healthy control |
| `tidwall/gjson` | `./...` | `validation-library-control` | healthy control |
| `prometheus/prometheus` | `./model/labels` | `medium-service-scoped` | `coverage_prefilter=true` |
| `kubernetes/apimachinery` | `./pkg/api/resource` | `large-multipackage-scoped-retargeted` | `coverage_prefilter=true`, resolved Go `1.26.0` |

The broader targets were not picked ad hoc. They were the current best
released-surface candidates after:

- broader-profile retargeting away from the original under-signal
  `apimachinery ./pkg/util/sets` sample in `#265`
- hosted per-repository toolchain resolution in `#269`, which removed the
  hidden baseline failure that had previously collapsed hosted `apimachinery`
  runs to zero effective work

## Result

Aggregate result from run `27848829038`:

- selected repos: `4`
- reports captured: `4`
- missing reports: `0`
- generated mutants: `40`
- covered mutants: `37`
- executed mutants: `37`
- effective mutants: `37`
- killed: `28`
- survived: `9`
- not covered: `3`
- timed out: `0`
- compile errors: `0`
- actionable review units: `8`
- semantic group review units: `2`
- semantic groups formed: `2`
- recommendation entries: `9`
- recommendation review units: `8`
- collapsed recommendation duplicates: `1`
- ledger entries: `6`
- governance suggestions: `10`
- repos with denominator warnings: `0`
- repos with reported failures: `0`

Per-repo result:

| Repo | Profile | Generated | Effective | Killed | Survived | Not covered | Actionable review units | Denominator health |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| `pflag-root` | `small-library-control` | `10` | `10` | `7` | `3` | `0` | `3` | `healthy` |
| `gjson-root` | `validation-library-control` | `10` | `10` | `7` | `3` | `0` | `2` | `healthy` |
| `prometheus-labels` | `medium-service-scoped` | `10` | `7` | `7` | `0` | `3` | `0` | `healthy` |
| `apimachinery-resource` | `large-multipackage-scoped-retargeted` | `10` | `10` | `7` | `3` | `0` | `3` | `healthy` |

## Interpretation

### Smaller control profiles remain stable

The small-library and validation-library controls stayed healthy and
review-bearing:

- `pflag-root` again produced `3` actionable review units
- `gjson-root` again produced `2` actionable review units after semantic
  collapse

That means the current public release did not lose the denominator-healthy
signal already seen in the bounded control sample.

### Medium service profile now contributes bounded effective work

`prometheus-labels` is no longer a denominator-collapse case in the released
sample:

- `generated=10`
- `effective=7`
- `killed=7`
- `not_covered=3`
- denominator health stayed `healthy`

It still did not produce survivors in this bounded run, but it now contributes
real effective mutation work rather than only a `no_effective_mutants`
warning.

### Large multipackage profile is now healthy on the released path

`apimachinery-resource` is the strongest release-surface change in this wave:

- `generated=10`
- `effective=10`
- `killed=7`
- `survived=3`
- `actionable_review_units=3`
- denominator health `healthy`
- resolved toolchain `go1.26.0`

That closes the earlier broader-profile failure mode where the large scoped
target contributed zero effective work on the hosted released path.

## Comparison To The Earlier Released v0.4.1 Profile Sample

This should not be read as a pure version-only benchmark. The candidate set
changed intentionally:

- `logrus-root` is no longer part of this broader-profile sample
- `apimachinery-resource` replaces the older `apimachinery-sets` target
- the workflow now resolves per-repository Go toolchains when target repos
  require a newer version than the action minimum

Even with that caveat, the current released sample is materially stronger:

| Released broader-profile sample | Selected repos | Effective | Killed | Survived | Not covered | Warning repos | Actionable review units |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| `v0.4.1` mixed-profile sample `27842880305` | `5` | `24` | `18` | `6` | `16` | `2` | `5` |
| `v0.4.2` calibrated sample `27848829038` | `4` | `37` | `28` | `9` | `3` | `0` | `8` |

That does not prove every broader-profile hosted rollout is solved. It does
prove that the current public release plus the calibrated target set now yields
substantially stronger mixed-profile evidence than the older `v0.4.1` sample.

## Findings

1. The released `v0.4.2` GitHub Action now has a credible broader-profile
   hosted sample, not only small-library controls.
2. The calibrated broader-profile targets from `#265` and `#269` transferred
   cleanly into the current public release:
   - `prometheus-labels` now contributes real effective work
   - `apimachinery-resource` now contributes survivors and actionable review
     units on the hosted released path
3. The aggregate released-surface review yield is materially stronger than the
   older `v0.4.1` broader sample:
   - `effective`: `24 -> 37`
   - `survived`: `6 -> 9`
   - `not_covered`: `16 -> 3`
   - `warning_repos`: `2 -> 0`
   - `actionable_review_units`: `5 -> 8`
4. This is better evidence for the maturity story than the older broader sample
   because it proves the newer review surfaces on the current public release
   across a healthier mixed-profile set, not just across the original
   small-library controls.
5. The remaining gap is no longer "can the broader-profile hosted sample
   produce any effective work at all?" The remaining gap is broader external
   maintainer adoption and repeated field evidence over time.

## Threats To Validity

1. The wave is still intentionally bounded to `10` mutants per repository with
   a `5m` budget, so it remains a CI-shaped adoption sample rather than a deep
   mutation campaign.
2. The broader-profile repos still use one scoped target each, not whole-repo
   breadth.
3. The sample is now healthier, but it is still curated. It should not be
   oversold as proof for every service-repo or monorepo rollout shape.
4. `prometheus-labels` still produced no survivors in this bounded pass, so the
   main signal there is denominator health and effective work rather than
   survivor triage depth.

## Conclusion

Issue `#288` now has the evidence it needed:

- the released `github-action@v0.4.2` hosted path was exercised across a
  healthier mixed-profile sample
- the smaller control repos remained stable
- both broader-profile targets now contribute effective mutation work on the
  released hosted path
- the current public release has stronger broader-profile evidence than the
  older `v0.4.1` sample and can support more precise maturity language
