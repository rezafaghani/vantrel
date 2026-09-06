# Go Raw Archive Pipeline

This package connects the ingestion SDK raw-capture hook to the MRX writer.

`rawarchive.Sink` implements `ingestion.RawSink` and writes each raw provider `Record` as an MRX frame while preserving the original payload bytes.

Run checks:

```sh
go test ./...
```
