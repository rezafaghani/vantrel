# Go Validation Engine

This package provides the first in-process validation engine.

It implements `ingestion.Validator` and returns structured validation findings for invalid canonical events.

Initial checks:

- event id, series id, type and payload are present
- JSON payload is parseable
- quote bid is less than or equal to ask
- actual and forecast values are numeric and finite

Run checks:

```sh
go test ./...
```
