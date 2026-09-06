# ADR 0002: Kafka Event Backbone

Status: Accepted

## Context

Market ingestion, validation, replay, storage, surveillance and analytics need ordered event movement and decoupled consumers.

## Decision

Use Apache Kafka as the platform event backbone.

Initial topic conventions are recorded in `contracts/kafka/topics.json`.

Topics are versioned with `.vN`, keyed explicitly and tied to a value contract. Replayable business streams must be recoverable from Kafka, while byte-perfect raw payload recovery remains the responsibility of MRX.

## Consequences

- Canonical events should be replayable from Kafka topics.
- Producers and consumers need schema compatibility and idempotency rules.
- Kafka is not installed during documentation-only architecture work.
- Topic additions and changes should update the registry and pass `make kafka-check`.
