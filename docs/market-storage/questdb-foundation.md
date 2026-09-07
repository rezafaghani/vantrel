# QuestDB Market Storage Foundation

QuestDB is Vantrel's intended durable store for market/time-series observations.

The initial Go boundary is `libs/go/marketstore`.

The initial QuestDB schema contract is `contracts/questdb/market_observations.sql`.

It defines:

- `ActualObservation`
- `ForecastObservation`
- `Quality`
- query windows by `series_id` and time range
- a minimal `Store` interface
- an in-memory implementation for tests

Rules:

- QuestDB stores observation values and timestamps, not business semantics.
- Series meaning, ownership, licensing, relationships and data products stay in the Catalog.
- Forecasts and actuals are separate observation models.
- Forecast rows carry `forecast_run_id`, `issued_at`, `target_time` and horizon.
- Actual rows carry event time and revision.

This task does not deploy QuestDB, add SQL schemas, write ILP clients, consume Kafka or persist data durably.
