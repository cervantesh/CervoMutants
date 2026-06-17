# Architecture Decision Records

This directory stores small, durable decisions that contributors should not
have to reconstruct from issues, PR threads, or chat history.

## When To Add An ADR

Add an ADR when a decision:

- changes package boundaries or ownership of responsibilities
- changes a default product doctrine
- changes or constrains a compatibility promise
- keeps a significant surface experimental on purpose

## Current ADRs

- [0001 Baseline-first adoption](0001-baseline-first-adoption.md)
- [0002 Report schema v1 stays additive](0002-report-schema-v1-additive.md)
- [0003 Daemon and worker stay experimental until versioned](0003-daemon-worker-stays-experimental.md)
- [0004 Semantic triage lives outside mutator generation](0004-semantic-triage-separation.md)

## Format

Use this structure:

1. Status
2. Context
3. Decision
4. Consequences

Keep ADRs short and concrete. They are not essays.
