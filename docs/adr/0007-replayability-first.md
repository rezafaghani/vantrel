# ADR 0007: Replayability-first

Status: Accepted

## Context

Market data, validation, enrichment, order-book state, surveillance analysis and AI decisions must be explainable and recoverable.

## Decision

Design important platform flows so they can be replayed where technically practical.

## Consequences

- Live traffic must not be the only way to reconstruct state.
- Raw inputs, canonical events, checkpoints and lineage IDs become first-class concerns.
- Some designs may trade extra storage for recoverability.
