# Vantrel

Vantrel is a Kubernetes-native market-data, trading, analytics, surveillance, machine-learning and AI intelligence platform for energy and financial markets.

The project is open-source-oriented but is in very early development. It is not production trading software, not investment advice, and not suitable for live trading or regulated operational use.

## Initial Scope

Vantrel will initially focus on:

- Energy and power markets
- Equities

The long-term platform direction includes:

- Go, C++, Python, TypeScript and React
- Kubernetes, Helm and later GitOps
- Apache Kafka and Apache Flink
- QuestDB for market/time-series observations
- TiDB for catalog, trading and compliance metadata domains
- Open-source object storage for raw archival
- Keycloak, Open Policy Agent and Apache APISIX
- OpenTelemetry, Prometheus, Grafana, Loki and Tempo
- Feast and MLflow for AI/ML workflows where appropriate

## Architecture Principles

Vantrel is designed around:

- Replayability and recoverability
- Byte-perfect preservation of raw provider payloads
- Separation of market observations from catalog semantics
- Deterministic behavior where appropriate
- Auditability, data lineage and operational transparency
- Strong contracts and independent service boundaries
- Validation before trust
- Governed AI access through explicit tools, not unrestricted database credentials
- Deterministic trading controls that AI cannot bypass

## Major Domains

- Ingestion framework
- MRX raw archival
- Canonical market contracts
- Series catalog and data products
- Market storage
- Validation
- Trading and order-book recovery
- Surveillance
- Data Intelligence
- AI/ML platform
- Security and authorization
- Observability
- Kubernetes operations

See [docs/architecture/platform-architecture.md](docs/architecture/platform-architecture.md) for the intended high-level platform architecture, [docs/adr](docs/adr) for architecture decisions, [docs/specs/mrx-v0.md](docs/specs/mrx-v0.md) for the draft MRX raw archive specification, [libs/go/mrx](libs/go/mrx) for the initial Go MRX library and CLI, [docs/kubernetes/local-development.md](docs/kubernetes/local-development.md) for the local Kubernetes foundation, [contracts/proto](contracts/proto) for initial canonical market contracts, and [docs/catalog/core.md](docs/catalog/core.md) for the Series Catalog core.

## Development Workflow

All development happens as small, independently reviewable pull requests.

For every task:

1. Update `main` from GitHub.
2. Verify the working tree is clean.
3. Create a new branch from updated `main`.
4. Implement only the current task.
5. Add or update tests and documentation as needed.
6. Run applicable tests, formatters and linters.
7. Review the diff.
8. Commit, push and open a pull request targeting `main`.
9. Stop until the pull request is reviewed and merged.

Do not merge your own pull request and do not start dependent work before the previous pull request is merged.

## Project Status

Vantrel is currently at repository bootstrap stage. No services, infrastructure stack, market adapters, trading systems or AI systems have been implemented yet.

## Safety And Provenance

This repository must not contain secrets, employer code, proprietary datasets, Centrica-specific source code, confidential concepts or licensed data that cannot be redistributed in a future open-source release.

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md) and [docs/licensing.md](docs/licensing.md).
