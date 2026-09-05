# Vantrel Platform Architecture

Vantrel is a Kubernetes-native platform for market data, trading, analytics, surveillance, machine learning and governed AI intelligence across energy/power markets and equities.

This document describes the intended ecosystem. It is not an implementation plan for this pull request.

## System Context

```mermaid
flowchart LR
    Providers[Market data providers\nREST, WebSocket, FIX, files, queues] --> Vantrel[Vantrel Platform]
    Users[Traders, quants, risk,\ncompliance, data engineers] --> Gateway[API gateway]
    Gateway --> Vantrel
    Vantrel --> ObjectStore[Open object storage\nMRX raw archive]
    Vantrel --> QuestDB[QuestDB\nmarket observations]
    Vantrel --> TiDB[TiDB\ndomain metadata]
    Vantrel --> Kafka[Kafka\nevent backbone]
    Vantrel --> ML[Feast and MLflow]
    Vantrel --> Obs[OpenTelemetry,\nPrometheus, Grafana,\nLoki, Tempo]
```

## Domain Decomposition

```mermaid
flowchart TB
    Ingestion[Ingestion]
    Raw[MRX Raw Archive]
    Contracts[Canonical Contracts]
    Catalog[Series Catalog]
    Validation[Validation]
    MarketStore[Market Storage]
    Trading[Trading]
    OrderBook[Order Book]
    Surveillance[Surveillance]
    DataIntel[Data Intelligence]
    AIML[AI/ML Platform]
    Security[Security]
    Observability[Observability]

    Ingestion --> Raw
    Ingestion --> Contracts
    Contracts --> Validation
    Validation --> MarketStore
    Catalog --> DataIntel
    MarketStore --> DataIntel
    MarketStore --> AIML
    Trading --> Surveillance
    OrderBook --> Trading
    Security --> Ingestion
    Security --> DataIntel
    Observability -. instruments .- Ingestion
    Observability -. instruments .- Trading
```

Primary domains:

- Ingestion captures external data and emits raw and canonical records.
- MRX preserves original provider payloads byte-for-byte.
- Canonical contracts define language-neutral events.
- Catalog owns series meaning, relationships and data products.
- Validation keeps invalid data traceable instead of silently dropping it.
- Market storage stores observations, not semantics.
- Trading and order-book services remain deterministic and replayable.
- Surveillance detects suspicious market and trading behavior with evidence.
- Data Intelligence exposes governed AI tools grounded in Catalog data.
- AI/ML provides models, features, experiments and recommendations.
- Security handles identity, authorization and policy enforcement.
- Observability exposes traces, metrics, logs and health.

## Data Architecture

```mermaid
flowchart LR
    RawPayload[Provider payload bytes] --> MRX[MRX archive]
    RawPayload --> Parser[Provider parser]
    Parser --> Canonical[Canonical event]
    Canonical --> Validation[Validation result]
    Validation -->|valid| Kafka[Kafka topic]
    Validation -->|invalid| Quarantine[Quarantine/DLQ]
    Kafka --> QuestDB[QuestDB observations]
    Catalog[Catalog metadata in TiDB] --> QuestDB
    Catalog --> DataProducts[Data products]
```

QuestDB stores time-series values and observations. It does not define what a series means.

TiDB stores transactional/domain metadata such as catalog, trading and compliance state. These domains may share an engine but must not be coupled into one schema.

Object storage stores raw MRX archives separately from normalized market storage.

## Ingestion Architecture

Live, historical and replay processing share the same core pipeline.

```mermaid
flowchart LR
    Source[Provider source] --> Capture[Raw capture]
    Capture --> Archive[MRX archive]
    Capture --> Parse[Parse]
    Parse --> Normalize[Normalize]
    Normalize --> Validate[Validate]
    Validate --> Publish[Publish]
    Publish --> Kafka[Kafka]
```

Provider adapters handle authentication, protocol, endpoint behavior, pagination, provider sequence handling and rate limits.

The ingestion runtime handles retries, backoff, circuit breaking, rate limiting infrastructure, backpressure, checkpointing, idempotency, raw archival, logs, metrics, traces, health, graceful shutdown, validation integration, normalization integration, Kafka publishing and quarantine.

## Raw Archival Architecture

MRX, Market Raw eXchange, is Vantrel's raw archival format.

```mermaid
flowchart TB
    File[MRX file] --> Header[Header\nmagic, version]
    File --> Frames[Independent frames]
    Frames --> Metadata[Provider/source metadata]
    Frames --> Payload[Original payload bytes]
    Frames --> Checksums[Frame and payload checksums]
    File --> Indexes[Timestamp and sequence indexes]
    File --> Footer[File checksum and close marker]
```

MRX must support byte-perfect recovery, replay, evidence, backloading, debugging and corruption detection. It does not replace Protobuf, Arrow, Parquet or QuestDB.

## Catalog Architecture

The Catalog owns data meaning.

```mermaid
erDiagram
    PROVIDER ||--o{ SOURCE : owns
    SOURCE ||--o{ SERIES : publishes
    SERIES ||--o{ SERIES_VERSION : versions
    SERIES ||--o{ SERIES_RELATIONSHIP : from
    SERIES ||--o{ SERIES_RELATIONSHIP : to
    DATA_PRODUCT ||--o{ DATA_PRODUCT_MEMBER : contains
    SERIES ||--o{ DATA_PRODUCT_MEMBER : included_in
```

Catalog entities include providers, sources, series, series versions, relationships, data products, data product members, purpose classifications and validation profile references.

Series definitions are versioned and auditable. Dynamic health belongs to Series Status, not static metadata.

## Forecast Vs Actual Model

Forecasts and realized observations are separate semantic categories.

```mermaid
flowchart LR
    ForecastRun[ForecastRun\nissued_at, model/source, version] --> ForecastPoint[Forecast observation\nseries_id, target_time, horizon, value]
    ActualSeries[Actual series] --> ActualPoint[Actual observation\nseries_id, event_time, value, revision]
    ForecastSeries[Forecast series] -->|FORECAST_OF| ActualSeries
    ForecastPoint --> Metrics[Error metrics\nMAE, RMSE, bias,\ndirectional accuracy, PnL impact]
    ActualPoint --> Metrics
```

Forecast points reference a coherent ForecastRun. Forecast series link to actual series through Catalog relationships.

## Data Intelligence Architecture

Data Intelligence is a governed domain for AI-assisted data understanding, discovery and dataset construction.

```mermaid
flowchart TB
    UserQuestion[Natural language request] --> Assistant[Data Intelligence]
    Assistant --> Tools[Governed tools]
    Tools --> Catalog[Catalog API]
    Tools --> Health[Series Status]
    Tools --> Samples[Sample data access]
    Tools --> Lineage[Lineage and relationships]
    Tools --> Audit[Tool call audit log]
```

Tools include `search_series`, `get_series`, `get_series_health`, `get_series_relationships`, `compare_series`, `get_series_coverage`, `get_quality_metrics`, `get_data_product`, `get_data_lineage`, `get_sample_data` and `build_candidate_dataset`.

AI recommendations must be grounded in series that exist in the Catalog and must respect authorization.

## AI/ML Architecture

```mermaid
flowchart LR
    MarketData[Validated market data] --> Features[Feast feature store]
    Catalog[Catalog metadata] --> Features
    Features --> Training[Training jobs]
    Training --> MLflow[MLflow experiments/models]
    MLflow --> Inference[Model inference]
    Inference --> Decisions[Recommendations and proposals]
    Decisions --> Controls[Deterministic validation,\nrisk and execution policy]
```

Traditional ML should be used where appropriate. LLMs do not replace numerical models for forecasting, volatility, anomaly detection, order-flow prediction or portfolio optimization.

## Trading Architecture

AI may analyze, predict, recommend, propose and explain. It may not bypass authentication, authorization, risk limits, portfolio constraints or execution controls.

```mermaid
flowchart LR
    Proposal[AI or human proposal] --> Validation[Deterministic validation]
    Validation --> Risk[Risk engine]
    Risk --> Policy[Execution policy]
    Policy --> Execution[Execution service]
    Execution --> Orders[Orders and trades]
    Orders --> Surveillance[Surveillance]
```

## Order-Book Recovery

The order book is planned for C++ and must be deterministic.

```mermaid
sequenceDiagram
    participant R as Recovery
    participant S as Snapshot Store
    participant J as Event Journal
    participant B as Order Book
    R->>S: load latest valid snapshot
    R->>B: restore snapshot
    R->>J: read events after snapshot sequence
    J-->>R: sequenced events
    R->>B: replay events
    R->>B: verify checksum
    B-->>R: healthy only if no gaps and checksum matches
```

A book with a sequence gap must not remain healthy.

## Security Architecture

```mermaid
flowchart LR
    User[User or machine identity] --> Keycloak[Keycloak]
    Keycloak --> APISIX[Apache APISIX]
    APISIX --> OPA[Open Policy Agent]
    OPA --> Services[Vantrel services]
    Services --> Audit[Audit log]
```

Roles should include Trader, Quant, Risk Manager, Compliance, Data Engineer, Administrator, Read Only and AI/machine identities.

## Observability Architecture

OpenTelemetry is the instrumentation standard.

```mermaid
flowchart LR
    Services[Vantrel services] --> OTel[OpenTelemetry]
    OTel --> Prometheus[Prometheus metrics]
    OTel --> Loki[Loki logs]
    OTel --> Tempo[Tempo traces]
    Prometheus --> Grafana[Grafana]
    Loki --> Grafana
    Tempo --> Grafana
```

Important IDs should propagate where appropriate: `trace_id`, `correlation_id`, `ingestion_id`, `raw_record_id`, `canonical_event_id`, `series_id`, `forecast_run_id`, `order_id` and `trade_id`.

## Kubernetes Architecture

Vantrel is Kubernetes-first. Docker Compose is not the primary development runtime.

```mermaid
flowchart TB
    Dev[Developer] --> Kind[kind or k3d]
    Kind --> Helm[Helm releases]
    Helm --> Namespaces[Namespaces by domain]
    Namespaces --> Services[Services]
    Services --> Health[Readiness and liveness]
    Services --> Config[Config and secrets]
    Services --> Persistence[Persistent volumes]
    GitOps[Future Argo CD] --> Helm
```

Local development should use kind or k3d, Helm conventions, health checks, resource requests/limits, graceful shutdown and documented configuration.
