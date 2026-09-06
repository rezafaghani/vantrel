# Catalog API

Small HTTP API for the Series Catalog core.

Supported endpoints:

- `GET /healthz`
- `GET /v1/series?q=wind`
- `POST /v1/series`
- `GET /v1/series/{series_id}`
- `PUT /v1/series/{series_id}`
- `GET /v1/series/{series_id}/versions`
- `POST /v1/relationships`
- `GET /v1/series/{series_id}/relationships`
- `POST /v1/providers`
- `POST /v1/sources`
- `POST /v1/data-products`
- `GET /v1/data-products/{product_id}`
- `POST /v1/data-products/{product_id}/members`

Mutating routes call an authorization hook. The current implementation provides the hook point only; Keycloak/OPA integration belongs in a later security task.

Run checks:

```sh
go test ./...
```
