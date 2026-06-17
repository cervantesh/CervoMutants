# Distribution Audit

Tracking issue: #162

Date: 2026-06-17

This audit checks whether packaging, release artifacts, installation guidance,
and onboarding assets form a coherent public distribution path.

## Audit Scope

The audit covered:

- release workflow target matrix
- published release availability
- binary install guidance
- first-party GitHub Action path
- quickstart and example-workflow onboarding

## Verified Results

### Release workflow

The release workflow currently builds archives for:

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

It also publishes `SHA256SUMS` and release notes.

Source of truth:

- [`.github/workflows/release.yml`](../.github/workflows/release.yml)
- [docs/releasing.md](releasing.md)

### Public install paths

The repository now documents:

- `go install` for the fastest cross-platform setup
- GitHub release archives for pinned binary installs
- source builds for local development and branch validation

Source of truth:

- [docs/install.md](install.md)

### First-party CI integration

The repository now ships a first-party GitHub Action for bounded
`cervomut run` lanes, and the maintained example workflows use that action.

Source of truth:

- [`action.yml`](../action.yml)
- [docs/github-action.md](github-action.md)
- [docs/github-review-workflow.md](github-review-workflow.md)
- [`examples/`](../examples)

### Onboarding

The shortest supported path from first install to a useful report is now
documented explicitly, instead of being spread across multiple pages.

Source of truth:

- [docs/quickstart.md](quickstart.md)
- [docs/adoption-guide.md](adoption-guide.md)

## Operational Check

As part of this audit, the latest public release `v0.3.0` was checked and its
expected binary artifacts plus `SHA256SUMS` were published so the documented
archive-install path is backed by a real release payload.

## Validation

The repo-level validation used for this audit:

- `go test ./...`
- YAML parse validation for:
  - `.github/workflows/*.yml`
  - `examples/*/.github/workflows/*.yml`
  - `action.yml`

## Current Limits

This audit does not claim:

- package-manager distribution such as Homebrew, Scoop, or apt
- automatic branch-aware install channel selection
- support for every CPU/OS combination beyond the documented release matrix

Those remain separate packaging or ecosystem decisions, not hidden promises of
the current public distribution path.
