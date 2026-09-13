# Go Market Store

This package defines the initial market observation storage boundary.

It keeps forecast, actual, trade and quote observations as separate types and exposes a small `Store` contract. `MemoryStore` exists for tests and early local use only.

QuestDB is the intended durable backend for market observations. `ILPWriter` encodes observations as QuestDB-compatible InfluxDB Line Protocol against an `io.Writer`, but this package does not open network connections or deploy QuestDB.

`Publisher` implements the ingestion SDK publisher hook for actual, forecast, trade and quote canonical events.

Run checks:

```sh
go test ./...
```
