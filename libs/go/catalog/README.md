# Go Series Catalog Core

This package contains the initial Series Catalog domain model.

It represents:

- providers and sources
- series and auditable series versions
- explicit series relationships
- data products and members
- validation and normalization profile references
- multi-purpose series classification

TiDB persistence is intentionally not implemented yet because database infrastructure is not in place. The package exposes a `Store` contract and includes `MemoryStore` for tests and early local usage.

Run checks:

```sh
go test ./...
```
