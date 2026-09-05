# ADR 0004: QuestDB Market Observations

Status: Accepted

## Context

Vantrel needs high-throughput time-series storage for market observations.

## Decision

Use QuestDB for market and time-series observations.

## Consequences

- QuestDB stores values and timestamps, not business semantics.
- Series meaning is resolved through the Catalog domain.
- PostgreSQL and ClickHouse are not part of the current storage direction.
