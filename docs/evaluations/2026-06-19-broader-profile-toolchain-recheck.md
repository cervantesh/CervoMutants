# Broader Profile Toolchain Recheck After Hosted Apimachinery Failure

## Scope

- Tracking issue: `#269`
- Manifest: [external-github-action-wave-broader-profile-toolchain-recheck.json](./external-github-action-wave-broader-profile-toolchain-recheck.json)
- Successful hosted run: `27844801501`
- Final action ref: `39dcb82564a7a67d60a517ddabe75ce735604755`

## Root Cause

The earlier hosted `apimachinery` failures were not a real `generated=0` discovery gap.
They were baseline runner failures hidden behind zero denominators.

The causal chain was:

1. The pinned `kubernetes/apimachinery` commit `6bdb38e7395b5518f3111f71bdf16152db13ad46` declares `go 1.26.0` in `go.mod`.
2. The hosted external wave had been forcing `go-version: 1.25.6`.
3. Hosted GitHub Actions also runs with `GOTOOLCHAIN=local`, so it does not auto-upgrade to a newer toolchain.
4. That made baseline `go test` fail before mutation, which surfaced in the report as `runner_error: baseline tests failed before mutation`.

Local scouting had looked healthy because this workstation uses `GOTOOLCHAIN=auto`, so the local Go toolchain can resolve a newer required toolchain automatically.

Direct Linux reproduction against the same target repo showed the real failure:

```text
go: go.mod requires go >= 1.26.0 (running go 1.25.6; GOTOOLCHAIN=local)
```

## Workflow Fix

`external-action-wave.yml` now resolves toolchains per repo and records them in the artifact summary:

- `go_version_action_min`: minimum Go version required to build the checked-out CervoMutants action source
- `go_version_target`: Go version declared by the target repository `go.mod`
- `go_version`: resolved version actually used for the run

The resolver uses:

```text
resolved_go = max(action_min_go, requested_go_or_target_go)
```

That avoids both failure modes:

- old target repos no longer force the action itself onto an unsupported old Go version
- newer target repos like `apimachinery` can rise above the action minimum when needed

## Intermediate Check

Hosted run `27844684840` validated half of the diagnosis:

- `apimachinery-resource` passed when the workflow resolved to `1.26.0`
- older control repos failed because the first implementation naively used the target repo `go` version, which dropped `gjson` and `pflag` to `1.12`

That run proved the toolchain mismatch diagnosis, but also showed that the resolver must respect the action's own minimum Go version.

## Final Hosted Validation

Hosted run `27844801501` re-ran the broader-profile control set with the max-version resolver and succeeded for all four repos.

| Repo | Target | action min | target go | resolved go | Generated | Effective | Killed | Survived | Not covered | Status |
| --- | --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| `spf13/pflag` | `./...` | `1.25.6` | `1.12` | `1.25.6` | `10` | `10` | `7` | `3` | `0` | success |
| `tidwall/gjson` | `./...` | `1.25.6` | `1.12` | `1.25.6` | `10` | `10` | `7` | `3` | `0` | success |
| `prometheus/prometheus` | `./model/labels` | `1.25.6` | `1.25.0` | `1.25.6` | `10` | `7` | `7` | `0` | `3` | success |
| `kubernetes/apimachinery` | `./pkg/api/resource` | `1.25.6` | `1.26.0` | `1.26.0` | `10` | `10` | `7` | `3` | `0` | success |

Aggregate summary from the final artifact:

- selected repos: `4`
- reports captured: `4`
- generated mutants: `40`
- effective mutants: `37`
- killed: `28`
- survived: `9`
- not covered: `3`
- repos with reported failures: `0`

The exact summary artifact is persisted at
[2026-06-19-external-github-action-wave-broader-profile-toolchain-recheck-summary.json](./2026-06-19-external-github-action-wave-broader-profile-toolchain-recheck-summary.json).

## Outcome

The hosted/local divergence in `#269` was a toolchain-resolution problem, not a mutation-engine discovery problem.

After the workflow fix:

- `kubernetes/apimachinery ./pkg/api/resource` is a valid hosted broader-profile target again
- the control repos remain healthy
- hosted wave artifacts now expose resolved toolchain evidence directly instead of leaving this class of failure implicit
