# Go Ingestion SDK

This package defines the initial reusable ingestion runtime.

It covers:

- `LIVE`, `HISTORICAL` and `REPLAY` modes
- provider adapter interface
- raw capture, parse, validate, publish and quarantine hooks
- checkpoint load/save boundary
- retry with optional exponential backoff
- fake adapter and memory checkpoint store for tests

It intentionally does not implement provider protocols, MRX persistence, Kafka publishing, validation rules, metrics or tracing yet. Those are separate focused tasks.

Run checks:

```sh
go test ./...
```
