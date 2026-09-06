# Ingestion SDK

Task 9 adds the initial Go ingestion SDK.

The SDK separates provider-specific acquisition from common ingestion flow:

```text
Source -> Raw Capture -> Parse -> Validate -> Publish
```

Provider adapters own authentication, connection, protocol, endpoint behavior, pagination, provider sequence handling and rate-limit specifics.

The runtime owns the shared pipeline shape:

- retries with exponential backoff
- checkpoint loading and saving
- raw capture hook
- parser hook
- validation hook
- publisher hook
- quarantine hook
- graceful stop through `context.Context`

## Modes

- `LIVE`
- `HISTORICAL`
- `REPLAY`

All modes use the same runtime and adapter contract.

## Boundaries

This task does not connect to Kafka, write MRX archives, define validation rules, add metrics/tracing or connect to a real provider.

The fake adapter exists to prove the runtime contract and give future provider adapters a small test target.
