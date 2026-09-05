# ADR 0006: Separate Catalog Domain

Status: Accepted

## Context

Market values alone do not explain what a series means, how it is licensed, where it came from or how it relates to other series.

## Decision

Create a separate Catalog domain for providers, sources, series, versions, purposes, relationships and data products.

## Consequences

- Time-series databases remain observation stores.
- Series definitions can be versioned and audited.
- AI recommendations must be grounded in Catalog records.
