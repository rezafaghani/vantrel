# ADR 0010: Order-book Snapshot Replay Recovery

Status: Accepted

## Context

Order books require deterministic recovery, sequence-gap detection and health verification.

## Decision

Recover order books by loading a valid snapshot, replaying subsequent journaled events, checking sequence continuity and verifying final state.

## Consequences

- A book with missing sequence data must not remain healthy.
- Snapshots, journals and checksums are required concepts.
- The implementation is deferred to a later C++ order-book task.
