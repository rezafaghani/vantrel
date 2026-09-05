# ADR 0005: TiDB Transactional Metadata

Status: Accepted

## Context

Catalog, trading and compliance domains need transactional metadata storage.

## Decision

Use TiDB where transactional/domain metadata storage is appropriate.

## Consequences

- Domains may share a database engine without sharing schemas or ownership.
- Catalog, trading and compliance boundaries remain explicit.
- TiDB infrastructure is not introduced until a focused implementation task needs it.
