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

Local development can start QuestDB with:

```sh
make questdb-up
```

Then run the local dashboard with:

```sh
make run-app
```

This local path creates the schema, writes synthetic market observations through QuestDB ILP-over-HTTP, and reads table counts through QuestDB's HTTP SQL API. Kafka and Kubernetes deployment remain future work.
