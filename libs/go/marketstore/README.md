# Go Market Store

This package defines the initial market observation storage boundary.

It keeps forecast and actual observations as separate types and exposes a small `Store` contract. `MemoryStore` exists for tests and early local use only.

QuestDB is the intended durable backend for market observations, but this task does not add a QuestDB driver or deployment.

Run checks:

```sh
go test ./...
```
