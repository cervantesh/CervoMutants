# Changelog

All notable changes to CervoMutants should be recorded here.

The format is intentionally simple:

- keep newest versions first
- add a `docs/upgrade-notes/<version>.md` file before tagging a release
- keep entries grounded in merged behavior, not roadmap intent

## [Unreleased]

- No unreleased entries yet.

## [v0.3.0] - 2026-06-17

### Changed

- Hardened Windows-native mutation execution with safer temp-root handling and
  more conservative worker defaults.
- Added runtime warning metadata so reports expose selected temp-root and
  Windows execution caveats.

### Added

- Large-repo CI slicing mode with deterministic sharding and per-run bounds.
- Survivor-ranking calibration follow-up that reduces noisy comparator-boundary
  competition in review queues.

### Verification

- `go test ./...`

## [v0.2.0] - 2026-06-17

### Added

- First-class resource statuses in reports, including `memory_killed`,
  `skipped_resource`, and `pending_budget`.
- Structured failure artifacts for `run` and `eval`, including correlation IDs,
  `failure-debug.json`, and panic recovery.
- `stopped_reason`, `last_completed_mutant`, and per-mutant
  `memory_peak_bytes` in report outputs.

### Verification

- `go test ./...`

## [v0.1.0] - 2026-06-17

### Added

- Initial GitHub release for CervoMutants.
- Apache License 2.0, `NOTICE`, and trademark guidance.

### Changed

- Aligned repository naming, module path, and documentation with the
  `CervoMutants` public project name.

### Verification

- `go test ./...`
