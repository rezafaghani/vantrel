# Validation Engine

The initial Validation Engine is a small Go package at `libs/go/validation`.

It implements the ingestion SDK `Validator` hook, so invalid canonical events can be quarantined by the existing ingestion runtime instead of being silently discarded.

Validation results contain:

- event id
- series id
- validity
- structured findings with code, category, severity and message

Initial categories:

- `STRUCTURAL`
- `SEMANTIC`

Initial rules:

- required event id, series id, type and payload
- parseable JSON payload
- quote `bid <= ask`
- numeric, finite actual/forecast observation values

This task does not add a validation service, Kafka publishing, persistence, cross-source checks, ML anomaly detection or Catalog-backed reference validation.
