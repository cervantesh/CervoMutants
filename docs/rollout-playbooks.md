# Rollout Playbooks By Repository Profile

Tracking issue: #191

This document turns the existing examples, adoption guidance, and supported CLI
workflows into repeatable rollout paths for three common repository shapes:

- compact libraries
- medium services
- large monorepos or large CI-heavy Go repositories

Use these playbooks after confirming fit with
[docs/adoption-guide.md](adoption-guide.md). They are not replacements for the
examples; they are the decision-complete rollout layer built on top of those
examples.

## Shared Rules For Every Playbook

These rules apply before choosing a repository profile:

1. Make sure `go test ./...` is already stable enough to trust as a baseline.
2. Start with baseline-first adoption rather than immediate raw score gating.
3. Keep PR mutation lanes bounded before adding broader nightlies or campaigns.
4. Preserve report artifacts instead of relying on memory or one-number score
   summaries.
5. Move to the next rollout stage only after the current one is understandable
   and trusted.

## Playbook 1: Compact Library

Use this when:

- `./...` is still a realistic mutation target
- a single PR lane is operationally affordable
- the team wants the lowest-friction first adoption path

Start from:

- [`examples/small-library`](../examples/small-library)

### Rollout path

1. Run discovery locally:

   ```powershell
   cervomut doctor
   cervomut run ./... --dry-run --out .cervomut/preview
   ```

2. Establish the first accepted baseline:

   ```powershell
   cervomut run ./... --policy ci-fast --budget 5m --out .cervomut/reports
   cervomut baseline update
   ```

3. Add the first PR mutation lane using the example workflow shape.
4. Review survivors through summary, JSON, JUnit, and GitHub summary outputs.
5. Stay on `ci-fast` until the team can explain which survivors are useful and
   which are noise.

### Exit criteria before growing the lane

- PR runtime is acceptable
- survivor review is understandable without custom scripts
- the baseline is trusted enough to catch regressions without causing churn

## Playbook 2: Medium Service

Use this when:

- the repository has multiple internal packages or bounded domains
- PR feedback needs richer review output than a tiny library workflow
- a nightly lane is acceptable in addition to PR checks

Start from:

- [`examples/medium-service`](../examples/medium-service)

### Rollout path

1. Keep the first local dry run and `doctor` checks.
2. Add a bounded PR lane:

   ```powershell
   cervomut run ./... --policy ci-balanced --budget 10m --coverage-prefilter --report summary,json,junit,github-summary --out .cervomut/pr
   ```

3. Accept the baseline before introducing stronger policy expectations.
4. Add a nightly lane only after the PR lane is already trusted:

   ```powershell
   cervomut run ./... --policy nightly --budget 20m --report summary,json,junit,html,sarif,github-summary --out .cervomut/nightly
   ```

5. Use the richer outputs to review:
   - HTML survivor workbench
   - SARIF or GitHub-native findings
   - actionable score and survivor recommendations
6. Introduce quarantine or ownership routing only after the basic PR and nightly
   flows already make sense.

### Exit criteria before moving to larger-scale workflows

- PR lane is reviewable and not noisy
- nightly artifacts are preserved and used in practice
- the team can separate PR feedback from deeper nightly review expectations

## Playbook 3: Large Monorepo Or Large CI-Heavy Repository

Use this when:

- a single full mutation lane is too slow or too noisy
- deterministic sharding is required before broad coverage is viable
- different lanes need different targets, shard density, or file caps

Start from:

- [`examples/large-repo-ci`](../examples/large-repo-ci)

### Rollout path

1. Do not start with broad `./...` mutation in one lane.
2. Establish a bounded PR shard:

   ```powershell
   cervomut run ./... --policy ci-fast --slice-by package --shard 1/4 --max-files-per-run 20 --sample deterministic --out .cervomut/pr
   ```

3. Accept the baseline for the bounded PR shape first.
4. Add a wider nightly shard set only after the PR shard is stable:

   ```powershell
   cervomut run ./... --policy nightly --slice-by file --shard 3/12 --max-mutants-per-package 25 --sample deterministic --report summary,json,junit,html,sarif,github-summary --out .cervomut/nightly
   ```

5. Add campaign-style targeted runs only for bounded domains with clear value:

   ```powershell
   cervomut run ./pkg/catalog ./pkg/pricing --policy campaign --workers 2 --out .cervomut/campaign
   ```

6. Preserve every lane's outputs separately instead of trying to merge unrelated
   history automatically.
7. Treat daemon/worker as experimental unless and until its support status
   changes publicly.

### Exit criteria before claiming broader rollout maturity

- shards are deterministic and operationally understandable
- preserved artifacts exist per lane or domain
- the team can explain why a package, file, or shard is in scope
- runtime and review density stay bounded enough for CI

## Choosing The Next Step

When a rollout feels stuck, use this rule:

- If the problem is noise, narrow policy, target, or scope before adding more
  automation.
- If the problem is missing historical context, preserve artifacts and add
  history review before broadening mutation breadth.
- If the problem is raw runtime, shard or slice before adding more mutants.

## Related Guides

- [docs/adoption-guide.md](adoption-guide.md)
- [docs/example-repos.md](example-repos.md)
- [docs/feedback-intake.md](feedback-intake.md)
- [docs/maintainer-operations-pack.md](maintainer-operations-pack.md)
