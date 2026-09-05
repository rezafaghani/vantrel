# ADR 0008: Canonical Protobuf Contracts

Status: Accepted

## Context

Vantrel will use Go, C++, Python and TypeScript services that need shared event contracts.

## Decision

Use Protobuf for canonical language-neutral market contracts.

## Consequences

- Contracts need versioning and compatibility checks.
- Provider-specific payloads remain separate from canonical events.
- Initial contracts are deferred to Task 6.
