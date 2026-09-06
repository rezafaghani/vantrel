# Raw Archival Pipeline

The initial raw archival pipeline is a Go adapter between the ingestion SDK and the MRX writer.

`libs/go/rawarchive` provides `rawarchive.Sink`, which implements `ingestion.RawSink`.

For each ingestion `Record`, the sink writes one MRX frame containing:

- original provider payload bytes
- provider id
- source id
- source sequence when numeric
- receive timestamp
- source timestamp when present
- original encoding
- ingestion record metadata
- deterministic raw record id derived from provider, source, sequence and payload bytes

The sink currently uses MRX `NONE` compression because the Go MRX library has not implemented LZ4 or ZSTD yet.

This task does not add object storage, Kafka, replay orchestration, rollover, Kubernetes deployment or Catalog API integration.
