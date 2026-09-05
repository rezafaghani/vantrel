# ADR 0003: MRX Raw Archive

Status: Accepted

## Context

Provider payloads must be recoverable byte-for-byte for replay, evidence, debugging and lineage.

## Decision

Define Market Raw eXchange, MRX, as Vantrel's raw archival format for compressed framed provider payloads.

The draft v0 format is specified in [MRX v0 Specification](../specs/mrx-v0.md).

## Consequences

- Raw archival is independent from normalized market storage.
- MRX does not replace Protobuf, Arrow, Parquet or QuestDB.
- Implementation is deferred to a focused language/library task.
