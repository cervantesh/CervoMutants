# ADR 0002: Report Schema v1 Stays Additive

## Status

Accepted

## Context

The JSON report is consumed by CI, automation, and human review tooling. Once
public, silent field removals or semantic shifts would break downstream
consumers more severely than many internal code changes.

## Decision

As long as the report stays on `schema_version: "1"`, changes must be additive.

For schema v1:

- new fields must be optional for older consumers
- existing field meanings must stay stable
- breaking changes require a new schema version and explicit migration notes

## Consequences

- contributors must treat report changes as contract work, not just rendering
  work
- report-related tests and fixtures should be updated together with schema
  changes
- compatibility policy and release notes must be updated when a schema change
  affects consumers
