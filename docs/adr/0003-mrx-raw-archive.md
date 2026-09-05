# ADR 0003: MRX Raw Archive

Status: Accepted

## Context

Provider payloads must be recoverable byte-for-byte for replay, evidence, debugging and lineage.

## Decision

Define Market Raw eXchange, MRX, as Vantrel's raw archival format for compressed framed provider payloads.

## Consequences

- Raw archival is independent from normalized market storage.
- MRX does not replace Protobuf, Arrow, Parquet or QuestDB.
- A detailed MRX v0 specification is deferred to Task 2.
