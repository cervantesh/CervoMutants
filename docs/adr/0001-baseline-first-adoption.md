# ADR 0001: Baseline-First Adoption

## Status

Accepted

## Context

Mutation testing is valuable, but raw fail-under gating is too blunt for many
existing Go repositories. For a new public tool, immediate hard gating would
push users toward score gaming, suppressions, or abandonment before they build
trust in the signal.

## Decision

CervoMutants defaults to baseline-first adoption rather than immediate raw-score
enforcement.

That means:

- baseline comparison is a first-class workflow
- new survivors and regressions matter more than a single raw threshold on day
  one
- quarantine and suppression stay explicit and auditable instead of silently
  hiding noise

## Consequences

- product UX emphasizes comparison, governance, and actionable review over
  simple score chasing
- contributors should be cautious about changes that make baseline flows
  weaker, harder to audit, or less visible
- documentation and CI guidance should continue treating baseline workflows as
  the default adoption path
