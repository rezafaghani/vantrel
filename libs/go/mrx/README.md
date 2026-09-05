# Go MRX Library

This package is the initial Go reference implementation for MRX v0.

Implemented:

- MRX v0.1 file header
- uncompressed `NONE` frames
- writer close footer
- CRC32C header, payload and file validation
- whole-archive reader with recovered frame index
- frame lookup by frame number
- recoverable missing-footer reads
- typed unsupported-compression errors

Not implemented yet:

- LZ4 and ZSTD compression
- streaming reads for large archives
- serialized optional JSON index
- sidecar indexes

Run checks:

```sh
go test ./...
```

Run the CLI:

```sh
go run ./cmd/mrx inspect archive.mrx
go run ./cmd/mrx verify archive.mrx
go run ./cmd/mrx extract -frame 0 archive.mrx
go run ./cmd/mrx benchmark -frames 1000 -size 256
```
