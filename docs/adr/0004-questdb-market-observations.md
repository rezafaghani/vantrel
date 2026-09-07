# ADR 0004: QuestDB Market Observations

Status: Accepted

## Context

Vantrel needs high-throughput time-series storage for market observations.

## Decision

Use QuestDB for market and time-series observations.

Initial code defines the storage boundary in `libs/go/marketstore`. Durable QuestDB persistence will implement that boundary later.

The initial table contract is `contracts/questdb/market_observations.sql`.

## Consequences

- QuestDB stores values and timestamps, not business semantics.
- Series meaning is resolved through the Catalog domain.
- PostgreSQL and ClickHouse are not part of the current storage direction.
- Forecast and actual observations remain separate at the storage boundary.
- Schema changes should pass `make questdb-schema-check`.
