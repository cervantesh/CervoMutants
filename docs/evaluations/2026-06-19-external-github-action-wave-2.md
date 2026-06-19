# External GitHub Action Wave 2

Tracking issue: #210

Date: 2026-06-19

This document records the second external adoption wave for CervoMutants using
the first-party GitHub Action path on GitHub-hosted runners instead of local
CLI orchestration alone.

The goal of this wave was not to prove broad ecosystem adoption. The goal was
to prove that the current action path can execute end-to-end on real public Go
repositories, produce auditable artifacts, and surface the next real friction
items from hosted execution.

Preflight evidence for the local `act` path lives in
[2026-06-19-external-github-action-wave-2-preflight.md](2026-06-19-external-github-action-wave-2-preflight.md).

Committed aggregate artifact:

- [2026-06-19-external-github-action-wave-2-summary.json](2026-06-19-external-github-action-wave-2-summary.json)

## Wave Shape

The hosted wave ran through:

- [.github/workflows/external-action-wave.yml](../../.github/workflows/external-action-wave.yml)
- [external-github-action-wave-2-candidates.json](external-github-action-wave-2-candidates.json)

Exact GitHub Actions run:

- run id: `27831153375`
- run URL:
  `https://github.com/cervantesh/cervo-mutants/actions/runs/27831153375`
- branch used for bootstrap run: `codex/210-gh-hosted-wave-run`
- workflow result: `success`
- created at: `2026-06-19T14:19:44Z`
- updated at: `2026-06-19T14:20:46Z`

The bootstrap run initially needed a narrow branch push trigger because
`gh workflow run ... --ref codex/210-gh-hosted-wave-run` could not dispatch a
workflow file that only existed off the default branch yet. The committed
workflow state for merge keeps `workflow_dispatch` and drops that bootstrap-only
push trigger.

## Repositories

| Repo | Target | Profile | Policy | Budget | Max mutants |
| --- | --- | --- | --- | --- | ---: |
| `spf13/cobra` | `./doc` | small library | `ci-fast` | `5m` | `5` |
| `sirupsen/logrus` | `./...` | medium library | `ci-fast` | `5m` | `5` |
| `grpc/grpc-go` | `./status` | scoped large CI-heavy | `ci-fast` | `5m` | `5` |

All jobs ran on GitHub-hosted `ubuntu-latest` runners with the workflow
checking out:

1. this repository as the action source
2. the external target repository under a separate path
3. the first-party composite action via `uses: ./cervomut-source`

## Result

Artifacts produced by the run:

- `external-wave-summary`
- `external-wave-cobra-doc`
- `external-wave-logrus-root`
- `external-wave-grpc-status`

Aggregate result:

- selected repos: `3`
- reports captured: `3`
- missing reports: `0`
- killed: `0`
- survived: `0`
- not covered: `10`
- timed out: `0`
- compile errors: `0`

Per-repo result:

| Repo | Job | Report | Total | Not covered | Score | Actionable | Denominator health |
| --- | --- | --- | ---: | ---: | ---: | ---: | --- |
| `cobra-doc` | `success` | `full` | `5` | `5` | `0` | `0` | `warning: no_effective_mutants` |
| `grpc-status` | `success` | `full` | `0` | `0` | `0` | `0` | `healthy` |
| `logrus-root` | `success` | `full` | `5` | `5` | `0` | `0` | `warning: no_effective_mutants` |

## Findings

1. The first-party GitHub Action path works end-to-end on GitHub-hosted
   runners across multiple external repositories. The workflow planned, cloned
   targets, executed the composite action, captured reports, and uploaded
   artifacts without repo-specific patches.
2. This wave is meaningful operationally even though the mutation signal was
   weak. It proves the hosted execution surface is real and reproducible, which
   was the primary product claim under test in `#210`.
3. Hosted execution signal quality still needs calibration. Two repos produced
   only `not_covered` mutants with denominator warnings
   (`no_effective_mutants`), while `grpc-status` produced `0` generated
   mutants under the current bounded `ci-fast` settings.
4. The current friction is now concrete enough to drive follow-up work instead
   of hand-wavy concern. Signal-quality follow-up belongs under
   [#211](https://github.com/cervantesh/cervo-mutants/issues/211).
5. Workflow hygiene follow-up is also concrete. The run completed with GitHub
   deprecation annotations for artifact actions that still target Node 20, now
   tracked in [#220](https://github.com/cervantesh/cervo-mutants/issues/220).

## Hosted Runner Warnings

GitHub emitted these exact annotations during run `27831153375`:

- `actions/upload-artifact@v4` in each matrix `wave` job
- `actions/download-artifact@v5` and `actions/upload-artifact@v4` in the
  `summarize` job

These warnings did not block execution, but they are release-era workflow debt
and should be removed before using future hosted-wave evidence as a cleaner
public trust signal.

## Threats To Validity

1. This wave validates the hosted GitHub Action path more strongly than it
   validates mutation-score usefulness on these exact settings.
2. The run was intentionally bounded to `5` mutants per repository and used
   `ci-fast`, so it is a smoke-style hosted adoption pass, not a deep mutation
   campaign.
3. The candidate set covered multiple repository shapes, but only three repos
   were exercised in this wave.
4. The hosted run used Linux runners. It does not replace the broader support
   matrix already documented elsewhere.

## Conclusion

This wave closes the narrower gap around hosted execution:

- the first-party GitHub Action runs successfully on GitHub-hosted runners
- the workflow path is reproducible from committed artifacts
- the next friction items are now explicit and issue-tracked

It does not yet justify a strong claim that the current hosted defaults produce
high-signal mutation results on representative public repositories. That next
calibration step is the point of the follow-up work, not something this wave
should pretend to have already solved.
