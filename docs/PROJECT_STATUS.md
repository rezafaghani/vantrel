# Project Status

Vantrel is in early development.

No production services, real provider adapters, full infrastructure deployments, trading systems, surveillance systems, ML systems or AI agents have been implemented.

Current foundation:

- Establish project identity and contribution rules.
- Document the mandatory small-PR workflow.
- Document high-level architecture and ADRs.
- Specify MRX v0 and provide the initial Go MRX library and CLI.
- Define initial canonical market Protobuf contracts.
- Provide initial Series Catalog core/API packages.
- Provide an initial Go ingestion SDK.
- Provide a deterministic synthetic market source for SDK and catalog integration tests.
- Provide an initial ingestion raw-capture sink backed by MRX.
- Provide initial Kafka topic registry conventions without deploying Kafka.
- Provide an initial in-process validation engine.
- Provide an initial market observation storage boundary without deploying QuestDB.
- Provide an initial QuestDB observation schema contract without deploying QuestDB.
- Provide initial QuestDB ILP line encoding without opening database connections.
- Provide initial ingestion-to-marketstore mapping for actual, forecast, trade and quote observations.
- Provide a minimal local pipeline runner without deploying Kafka, QuestDB or Kubernetes.
