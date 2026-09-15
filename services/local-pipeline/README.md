# Local Pipeline

Small runnable pipeline for local development.

It runs the synthetic market source through:

- MRX raw capture
- synthetic parsing
- validation
- marketstore ILP encoding

It does not require Kafka, QuestDB or Kubernetes.

Run:

```sh
go run . -steps 3
```
