# Synthetic Market Source

Task 10 adds a deterministic in-memory provider under `libs/go/synthetic`.

The source is intentionally small and exists to prove the ingestion SDK can process reproducible provider data before any real market integration exists.

Generated records:

- `trade`: Apple trade price observations
- `quote`: Apple bid/ask observations
- `forecast`: DK1 wind production forecasts
- `actual`: DK1 realized wind production observations

Each raw record is JSON and carries provider id, source id, source sequence, timestamps, original encoding and series metadata through the ingestion SDK `Record`.

Catalog registration creates:

- provider `synthetic`
- source `synthetic-market-source`
- four logical series with stable Vantrel `series_id` values
- a `FORECAST_OF` relationship from DK1 wind forecast to DK1 wind actual
- data product `SYN_DK1_POWER_AND_APPLE_MARKET_SAMPLE`

The source does not publish to Kafka, write MRX archives or simulate a full exchange. Those remain later pipeline tasks.
