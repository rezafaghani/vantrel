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

QuestDB-backed dashboard:

```sh
make questdb-up
make run-app
```

Open `http://localhost:8080` and click `Run Synthetic Batch`.

Endpoints:

- `GET /`
- `GET /api/status`
- `GET /api/latest`
- `POST /api/run?steps=3`
