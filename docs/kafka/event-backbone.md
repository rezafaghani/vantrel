# Kafka Event Backbone

Kafka is Vantrel's planned event backbone for durable, replayable movement of canonical market data and domain events.

The initial convention is captured in `contracts/kafka/topics.json` and checked by `tools/check_kafka_topics.py`.

Topic rules:

- Topic names are versioned with a `.vN` suffix.
- Topic names use the `vantrel.<domain>.<stream>.vN` shape.
- Message keys must be explicit and stable for ordering-sensitive consumers.
- Value contracts must reference an existing contract or clearly name a future contract.
- Replayable business data uses finite retention unless the stream is compacted state.
- Kafka is not used for byte-perfect raw provider archival; MRX owns raw payload recovery.

Initial topics:

- `vantrel.market.canonical.v1`
- `vantrel.market.validation.v1`
- `vantrel.market.quarantine.v1`
- `vantrel.catalog.series.v1`
- `vantrel.ingestion.status.v1`

This task does not install Kafka, Strimzi, Redpanda, brokers, schema registries or client libraries.
