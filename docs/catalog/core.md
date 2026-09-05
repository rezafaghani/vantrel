# Series Catalog Core

The Series Catalog owns the meaning, origin, purpose and relationships of logical market and fundamental series.

Task 7 adds the initial Go domain model and persistence boundary. It does not add an API or TiDB schema.

## Initial Model

- `Provider`: external or internal provider identity.
- `Source`: feed, endpoint or protocol source owned by a provider.
- `Series`: stable logical series identified by `series_id`.
- `SeriesVersion`: immutable audit snapshot for each series update.
- `SeriesRelationship`: explicit relationship between two series.
- `DataProduct`: business grouping of series.
- `DataProductMember`: membership and role of a series in a data product.

Supported purposes:

- `PRE_TRADE`
- `REALTIME`
- `POST_TRADE`
- `SETTLEMENT`
- `REFERENCE`
- `TRAINING`
- `VALIDATION`
- `COMPLIANCE`

Supported relationship types:

- `FORECAST_OF`
- `ACTUAL_OF`
- `DERIVED_FROM`
- `AGGREGATED_FROM`
- `INPUT_TO`
- `OUTPUT_OF`
- `RELATED_TO`
- `CORRELATED_WITH`
- `REPLACED_BY`
- `INFLUENCED_BY`

## Persistence

TiDB is the intended transactional storage engine for Catalog data, but the Kubernetes/database foundation is not ready yet.

The Go package defines a `Store` interface and an in-memory implementation. A future TiDB repository should implement the same contract or replace it through an ADR if the contract proves too narrow.

## Boundaries

- Catalog defines series meaning; QuestDB stores observations.
- Forecast and actual series are separate and linked with relationships.
- Runtime systems should avoid expensive graph exploration on every market event.
- AI recommendations must be grounded in Catalog records and later include Series Status health.
