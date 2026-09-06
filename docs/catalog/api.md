# Catalog API

Task 8 adds a small HTTP API over the Series Catalog core.

It supports:

- series search
- series get
- series create and edit/version
- relationship create and lookup
- provider and source create
- data product create, get and member add

The API uses the Catalog package validation rules and exposes an authorization hook for mutating routes. Full authentication and OPA policy enforcement are deferred until the security foundation exists.

Persistence still uses the Catalog `Store` interface. TiDB-backed persistence is not implemented in this task.
