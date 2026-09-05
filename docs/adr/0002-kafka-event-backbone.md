# ADR 0002: Kafka Event Backbone

Status: Accepted

## Context

Market ingestion, validation, replay, storage, surveillance and analytics need ordered event movement and decoupled consumers.

## Decision

Use Apache Kafka as the platform event backbone.

## Consequences

- Canonical events should be replayable from Kafka topics.
- Producers and consumers need schema compatibility and idempotency rules.
- Kafka is not installed during documentation-only architecture work.
