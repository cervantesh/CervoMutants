# ADR 0003: Daemon And Worker Stay Experimental Until Versioned

## Status

Accepted

## Context

The repository contains a JSON-lines daemon/worker path, but it does not yet
provide version negotiation, leases, retries, or a supported multi-worker
coordination story.

Treating that path as a stable distributed-execution surface would overpromise
and create compatibility obligations the code does not yet meet.

## Decision

`cervomut daemon` and `cervomut worker` remain experimental until the protocol
is explicitly versioned and promoted.

That means:

- explicit opt-in is required
- the protocol does not yet carry backward-compatibility guarantees
- release and compatibility docs must keep this surface out of the supported
  contract set

## Consequences

- contributors must not market daemon/worker as production-ready distribution
- any proposal to graduate the protocol should come with a new ADR or equivalent
  explicit contract change
- experimental protocol changes can move faster than supported CLI or report
  surfaces, but they still need clear docs
