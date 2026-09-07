# QuestDB Observation Schema

`contracts/questdb/market_observations.sql` defines the first QuestDB table contract for market observations.

Tables:

- `market_actual_observations`
- `market_forecast_observations`

The tables intentionally duplicate only runtime observation identifiers and values:

- `series_id`
- observation timestamps
- numeric value and unit
- quality fields
- lineage ids such as `raw_record_id`, `canonical_event_id` and `ingestion_id`

They do not include Catalog-owned semantics such as display names, descriptions, licenses, owners, relationships or data products.

Forecast rows use `target_time` as the designated timestamp and retain `forecast_run_id`, `issued_at` and `horizon_seconds`.

Actual rows use `event_time` as the designated timestamp and retain `revision`.

Run `make questdb-schema-check` after changing the schema.

`libs/go/marketstore.ILPWriter` writes rows using QuestDB-compatible InfluxDB Line Protocol. The writer targets an `io.Writer`; connection management is left to a later adapter task.
