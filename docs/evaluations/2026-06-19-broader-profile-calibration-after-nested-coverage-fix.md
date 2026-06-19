# 2026-06-19 Broader-Profile Calibration After Nested Coverage Fix

Tracking issue: [#265](https://github.com/cervantesh/cervo-mutants/issues/265)

Follow-up issue: [#269](https://github.com/cervantesh/cervo-mutants/issues/269)

Date: 2026-06-19

This note records the first broader-profile hosted calibration pass after the
nested package coverage-profile fix from [#267](https://github.com/cervantesh/cervo-mutants/issues/267)
merged in [#268](https://github.com/cervantesh/cervo-mutants/pull/268).

The goal here was narrower than "solve every broader profile." The goal was to
separate what was fixed by the nested coverage-path bug from what still appears
to be a hosted-environment limitation.

## What Changed

Two product-surface changes landed on the calibration branch before the hosted
runs:

1. `external-action-wave.yml` now accepts per-job `coverage_prefilter` and
   passes it through to the GitHub Action.
2. Two hosted manifests were added:
   - `external-github-action-wave-broader-profile-calibration.json`
   - `external-github-action-wave-broader-profile-retargeted.json`

Committed aggregate artifacts:

- [2026-06-19-external-github-action-wave-broader-profile-calibration-summary.json](2026-06-19-external-github-action-wave-broader-profile-calibration-summary.json)
- [2026-06-19-external-github-action-wave-broader-profile-retargeted-summary.json](2026-06-19-external-github-action-wave-broader-profile-retargeted-summary.json)

## Inputs

Hosted runs on the calibration branch:

- run `27843959529`
  - manifest:
    `docs/evaluations/external-github-action-wave-broader-profile-calibration.json`
  - action ref: `4f02ce8456579fd59b51357511b53aff922bd627`
- run `27844128977`
  - manifest:
    `docs/evaluations/external-github-action-wave-broader-profile-retargeted.json`
  - action ref: `176f9e23ffbb0ac56b0b7fd15f62e94066bd889c`

Local scout evidence used to choose the retarget:

| Repo | Target | Generated | Effective | Killed | Survived | Not covered | Healthy |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `kubernetes/apimachinery` | `./pkg/util/sets` | `4` | `4` | `2` | `2` | `0` | `true` |
| `kubernetes/apimachinery` | `./pkg/api/resource` | `4` | `4` | `3` | `1` | `0` | `true` |
| `kubernetes/apimachinery` | `./pkg/labels` | `4` | `4` | `2` | `2` | `0` | `true` |
| `kubernetes/apimachinery` | `./pkg/runtime/schema` | `4` | `4` | `0` | `4` | `0` | `true` |

## Hosted Result

### First Hosted Calibration Wave

| Repo | Target | coverage_prefilter | Generated | Effective | Killed | Survived | Not covered | Healthy |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `pflag-root` | `./...` | `false` | `10` | `10` | `7` | `3` | `0` | `true` |
| `gjson-root` | `./...` | `false` | `10` | `10` | `7` | `3` | `0` | `true` |
| `prometheus-labels` | `./model/labels` | `true` | `10` | `7` | `7` | `0` | `3` | `true` |
| `apimachinery-sets` | `./pkg/util/sets` | `true` | `0` | `0` | `0` | `0` | `0` | `false` |

Aggregate:

- selected repos: `4`
- generated mutants: `30`
- effective mutants: `27`
- killed: `21`
- survived: `6`
- not covered: `3`
- warning repos: `0`

### Retargeted Hosted Wave

The second wave replaced `apimachinery-sets` with `apimachinery-resource`
because the local scout made `./pkg/api/resource` the strongest alternative.

| Repo | Target | coverage_prefilter | Generated | Effective | Killed | Survived | Not covered | Healthy |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `pflag-root` | `./...` | `false` | `10` | `10` | `7` | `3` | `0` | `true` |
| `gjson-root` | `./...` | `false` | `10` | `10` | `7` | `3` | `0` | `true` |
| `prometheus-labels` | `./model/labels` | `true` | `10` | `7` | `7` | `0` | `3` | `true` |
| `apimachinery-resource` | `./pkg/api/resource` | `true` | `0` | `0` | `0` | `0` | `0` | `false` |

The aggregate stayed effectively unchanged because the hosted `apimachinery`
job still contributed zero mutants after retargeting.

## Findings

1. The nested coverage-path fix plus hosted `coverage_prefilter` wiring are
   enough to recover real signal for the Prometheus package target.
   `prometheus-labels` moved from the previous `effective=0` shape to:
   - `generated=10`
   - `effective=7`
   - `killed=7`
   - `not_covered=3`
   - `healthy=true`
2. The large-multipackage `apimachinery` profile is no longer best explained by
   a bad target alone. Two different hosted package targets still produced
   `generated=0`, while the same branch and same mutation engine produced
   healthy local scout signal for multiple `apimachinery` package targets.
3. That means the remaining `apimachinery` gap is now narrow and explicit:
   hosted package-target generation or discovery for this repo still diverges
   from the local result. That follow-up belongs in [#269](https://github.com/cervantesh/cervo-mutants/issues/269).
4. The smaller-library controls stayed healthy across both waves, so the new
   workflow wiring itself did not regress the known-good hosted path.

## Conclusion

`#265` now has a concrete, evidence-based path instead of a generic
"broader-profile under-signal" bucket:

- keep `prometheus ./model/labels` as a valid hosted broader-profile target
  under `coverage_prefilter`
- do not treat `kubernetes/apimachinery` as a working hosted broader-profile
  representative yet
- track the remaining hosted/local `apimachinery` divergence in `#269`

That is enough to narrow future hosted validation claims and keep the remaining
large-multipackage gap explicitly tracked instead of leaving it as unresolved
context.
