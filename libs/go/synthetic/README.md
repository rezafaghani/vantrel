# Go Synthetic Market Source

This package provides a deterministic synthetic provider for local development and ingestion SDK tests.

It emits raw JSON records for:

- Apple trades
- Apple quotes
- DK1 wind forecasts
- DK1 wind actuals

The source uses the Task 9 ingestion SDK and can register matching provider, source, series, forecast/actual relationship and data-product metadata in the Series Catalog.

Run checks:

```sh
go test ./...
```
