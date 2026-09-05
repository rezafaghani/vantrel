# ADR 0011: Governed AI Data Intelligence Tools

Status: Accepted

## Context

AI components need to understand Vantrel data without unrestricted database credentials or invented provider data.

## Decision

Expose governed Data Intelligence tools for catalog search, health, relationships, lineage, samples, quality metrics and candidate dataset construction.

## Consequences

- Authorization applies to AI tool calls.
- Tool calls must be auditable.
- Recommendations must be grounded in existing Catalog records.
