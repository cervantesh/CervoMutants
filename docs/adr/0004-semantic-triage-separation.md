# ADR 0004: Semantic Triage Lives Outside Mutator Generation

## Status

Accepted

## Context

Mutation generation and mutation interpretation are related but different
responsibilities.

If semantic actionability heuristics live inside mutation generation, the code
becomes harder to test, harder to evolve, and more likely to mix product policy
with AST rewriting mechanics.

## Decision

Semantic triage belongs in `pkg/triage`, while `pkg/mutator` remains focused on
generation and structural context capture.

That means:

- mutators generate candidates and attach structural metadata
- triage classifies actionability, equivalence risk, grouping, and similar
  review heuristics
- engine/report layers consume triage output instead of inventing parallel
  heuristics ad hoc

## Consequences

- new review heuristics should default to `pkg/triage`
- `pkg/mutator` should not grow product-ranking logic casually
- contributors should preserve the separation unless there is a strong,
  documented reason to change it
