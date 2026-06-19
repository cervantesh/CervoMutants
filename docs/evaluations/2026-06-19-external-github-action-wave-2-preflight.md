# External GitHub Action Wave 2 Preflight

Tracking issue: #210

Date: 2026-06-19

This document records the preflight work for the second external adoption wave.
Unlike the first public wave, this phase starts from the first-party GitHub
Action path instead of raw local CLI commands.

The goal of this preflight is not to claim the wave is complete. The goal is to
prove or disprove the local execution path early enough that the actual wave can
be run reproducibly and any runner-specific friction can be documented instead
of rediscovered.

## Candidate Repository Mix

The selected candidate set is stored in
[external-github-action-wave-2-candidates.json](external-github-action-wave-2-candidates.json).

Initial mix:

| Repo | Target | Profile | Why |
| --- | --- | --- | --- |
| `spf13/cobra` | `./doc` | small library | Keeps the first smoke run bounded while still using a real public repository. |
| `sirupsen/logrus` | `./...` | medium library | Reuses a stable medium-sized external repo from earlier validation material. |
| `grpc/grpc-go` | `./status` | scoped large CI-heavy repo | Adds a larger repository shape without making the first wave attempt unbounded. |

## Action Ref Under Test

The preflight intentionally pinned the action to the current repo-head commit
instead of `v0.4.0` so the test includes the post-release GitHub Actions
hardening merged on 2026-06-19:

- action ref: `f191acdf89869a4705f2d5ad63d5121cf32a7990`

That keeps the preflight aligned with current `main` rather than the last
release tag.

## Local Execution Path

Environment used for the preflight:

- host OS: Windows
- Docker engine: `29.4.3`
- local workflow runner: `act v0.2.89`
- Docker socket mode for `act`: `--container-daemon-socket -`
- explicit runner image for `act` dry-run: `catthehacker/ubuntu:act-latest`

The `act` binary was downloaded into a temporary directory outside the
repository so the repo state remains clean.

## Workflow Shape Tested

The first smoke workflow used this shape against `cobra`:

```yaml
name: cervomut-wave
on:
  workflow_dispatch:
jobs:
  mutation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - name: Run CervoMutants
        uses: cervantesh/cervo-mutants@f191acdf89869a4705f2d5ad63d5121cf32a7990
        with:
          go-version: "1.25.6"
          targets: ./doc
          policy: ci-fast
          budget: 5m
          report: summary,json,github-summary
          out: .cervomut/pr
          max-mutants: "5"
          workers: "1"
```

## Preflight Result

### 1. Dry-run succeeded

This command validated the workflow shape under `act` without executing the job:

```powershell
act workflow_dispatch `
  -W .github/workflows/cervomut-wave.yml `
  -j mutation `
  -n `
  -P ubuntu-latest=catthehacker/ubuntu:act-latest `
  --container-daemon-socket -
```

Result:

- workflow plan resolved successfully
- the remote action ref cloned correctly
- the composite action structure was understood by `act`

### 2. Real run failed in the local runner layer

The real `cobra` run exited with code `1` after about `85.696` seconds.

The verbose `act` log showed the actionable failure:

- composite action step failed with `exitcode '127': command not found`
- the action uses `shell: pwsh` in its composite steps
- the chosen local `act` medium image does not provide `pwsh`

This is runner friction, not yet evidence of a product regression in
CervoMutants itself.

## Meaning Of The Failure

This finding matters because it shows a concrete divergence between:

- the real GitHub-hosted `ubuntu-latest` environment, where `pwsh` is expected
  to exist for composite actions like this one
- the lighter local `act` image path, which is attractive for maintainers but
  is not environment-equivalent by default

That divergence is exactly the kind of adoption friction the second wave is
supposed to surface.

## Next Execution Options

The next realistic paths are:

1. rerun the wave with a fuller GitHub-runner-style `act` image that includes
   `pwsh`
2. run the wave through actual GitHub-hosted runners on disposable or dedicated
   public validation repositories

Additional preflight evidence:

- `docker manifest inspect catthehacker/ubuntu:full-latest` confirms a fuller
  local `act` runner image exists
- the first `docker run` attempt for that image exceeded a 15-minute local
  timeout on this host, so the fuller-image path is available but not yet
  proven lightweight enough for a quick maintainer loop

Until one of those is chosen, the wave should not be described as completed.

## Current Conclusion

The wave itself is still pending, but the preflight already established useful
evidence:

- candidate repos are selected
- the current action ref is pinned
- the local GitHub Action execution path was exercised
- the first concrete friction item is now identified and attributable
