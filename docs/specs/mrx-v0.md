# MRX v0 Specification

Status: Draft

MRX, Market Raw eXchange, is Vantrel's raw archival format for external provider payloads. It stores original bytes in compressed framed containers so payloads can be recovered byte-for-byte for replay, evidence, backloading, debugging and lineage.

MRX is not a canonical event format, analytics format or time-series database format. It does not replace Protobuf, Arrow, Parquet or QuestDB.

## Goals

- Preserve raw provider payload bytes exactly.
- Support independent frame validation and partial recovery.
- Keep enough metadata for replay and lineage.
- Allow forward-compatible readers.
- Be implementable in Go, C++ and Python without shared runtime dependencies.

## Non-goals

- Defining normalized market events.
- Storing catalog semantics.
- Replacing object storage indexes or data lake table formats.
- Encrypting payloads. Encryption belongs to storage and key-management layers.

## Byte Order

All fixed-width integers are little-endian.

Timestamps are signed 64-bit Unix nanoseconds unless otherwise stated.

Strings and metadata keys are UTF-8.

## File Layout

```text
+-------------+
| File header |
+-------------+
| Frame 0     |
+-------------+
| Frame 1     |
+-------------+
| ...         |
+-------------+
| Index       | optional in v0
+-------------+
| Footer      |
+-------------+
```

A reader MUST be able to scan frames without an index. A writer SHOULD emit a footer when closing cleanly.

## File Header

| Offset | Size | Type | Field | Description |
| --- | ---: | --- | --- | --- |
| 0 | 4 | bytes | magic | `MRX0` |
| 4 | 1 | uint8 | major_version | `0` |
| 5 | 1 | uint8 | minor_version | `1` |
| 6 | 2 | uint16 | header_length | Header bytes including fixed and variable fields |
| 8 | 4 | uint32 | flags | File flags |
| 12 | 16 | bytes | file_id | Writer-generated opaque ID |
| 28 | 8 | int64 | created_at_ns | File creation time |
| 36 | 4 | uint32 | metadata_length | Header metadata bytes |
| 40 | N | bytes | metadata | Canonical JSON metadata |

The fixed header length is 40 bytes. `header_length` MUST be at least 40 and MUST equal `40 + metadata_length` for v0.1 files.

File flags:

| Bit | Meaning |
| ---: | --- |
| 0 | Footer expected |
| 1-31 | Reserved, MUST be zero |

Header metadata is canonical JSON:

- Object keys sorted lexicographically.
- No insignificant whitespace.
- UTF-8 encoding.
- Integer timestamps encoded as JSON numbers.

Required header metadata keys:

- `creator`: writer name and version
- `provider_id`: provider identifier
- `source_id`: source/feed identifier

Optional header metadata keys:

- `market`
- `asset_class`
- `environment`
- `schema_hint`
- `notes`

## Frame Layout

Each frame is independently readable and validates its own payload.

| Offset | Size | Type | Field | Description |
| --- | ---: | --- | --- | --- |
| 0 | 4 | bytes | frame_magic | `MRXF` |
| 4 | 2 | uint16 | frame_header_length | Bytes from frame start to payload start |
| 6 | 1 | uint8 | frame_version | `0` |
| 7 | 1 | uint8 | compression_id | Compression algorithm |
| 8 | 4 | uint32 | frame_flags | Frame flags |
| 12 | 8 | uint64 | frame_number | Zero-based frame number |
| 20 | 16 | bytes | raw_record_id | Opaque record ID |
| 36 | 8 | int64 | receive_time_ns | Time Vantrel received the payload |
| 44 | 8 | int64 | source_time_ns | Provider/source event time, or `-1` if unknown |
| 52 | 8 | int64 | source_sequence | Provider sequence, or `-1` if unavailable |
| 60 | 4 | uint32 | metadata_length | Frame metadata bytes |
| 64 | 8 | uint64 | uncompressed_length | Raw payload byte length |
| 72 | 8 | uint64 | compressed_length | Stored payload byte length |
| 80 | 4 | uint32 | payload_crc32c | CRC32C of uncompressed payload |
| 84 | 4 | uint32 | frame_header_crc32c | CRC32C of frame header bytes before this field |
| 88 | M | bytes | metadata | Canonical JSON metadata |
| 88+M | P | bytes | payload | Compressed or uncompressed payload bytes |

The fixed frame header length is 88 bytes. `frame_header_length` MUST equal `88 + metadata_length` for v0.1 files.

Frame flags:

| Bit | Meaning |
| ---: | --- |
| 0 | Source sequence is present |
| 1 | Source timestamp is present |
| 2 | Payload is provider text |
| 3 | Payload is provider binary |
| 4 | Frame is replayed from another MRX archive |
| 5-31 | Reserved, MUST be zero |

Exactly one of bits 2 or 3 SHOULD be set when known.

Required frame metadata keys:

- `provider_id`
- `source_id`
- `encoding`

Optional frame metadata keys:

- `ingestion_id`
- `endpoint`
- `content_type`
- `source_partition`
- `source_offset`
- `request_id`
- `trace_id`
- `correlation_id`
- `schema_hint`

Unknown metadata keys MUST be preserved when rewriting frames unless the writer explicitly documents that it is normalizing metadata.

## Compression Identifiers

| ID | Name | Description |
| ---: | --- | --- |
| 0 | NONE | Payload is stored as original bytes |
| 1 | LZ4 | LZ4 block format |
| 2 | ZSTD | Zstandard frame format |
| 3-127 | Reserved | Future Vantrel algorithms |
| 128-255 | Private | Implementation-specific experiments |

Readers MUST support `NONE`. Readers MAY support `LZ4` and `ZSTD`. A reader that encounters an unsupported compression ID MUST report a typed unsupported-compression error without treating the file as corrupt.

## Checksums

MRX v0 uses CRC32C for fast corruption detection.

- `payload_crc32c` is computed over the uncompressed raw payload bytes.
- `frame_header_crc32c` is computed over bytes `[0, 84)` of the frame header.
- Footer `file_crc32c` is computed over all bytes before the footer magic.

A reader MUST validate the frame header checksum before trusting frame lengths. A reader MUST validate the payload checksum after decompression.

## Footer

The footer marks a cleanly closed file.

| Offset | Size | Type | Field | Description |
| --- | ---: | --- | --- | --- |
| 0 | 4 | bytes | footer_magic | `MRXE` |
| 4 | 8 | uint64 | frame_count | Number of complete frames |
| 12 | 8 | uint64 | index_offset | Offset to index, or `0` if absent |
| 20 | 8 | uint64 | index_length | Index byte length, or `0` if absent |
| 28 | 4 | uint32 | file_crc32c | CRC32C of bytes before footer |
| 32 | 4 | bytes | footer_end_magic | `DONE` |

A missing footer means the file may be truncated. Readers SHOULD recover complete frames up to the first invalid or incomplete frame and report the file as unclean.

## Index

Indexes are optional in v0. A reader MUST NOT require an index.

When present, the index is canonical JSON with this shape:

```json
{
  "version": 0,
  "frames": [
    {
      "frame_number": 0,
      "offset": 36,
      "receive_time_ns": 1735689600000000000,
      "source_time_ns": 1735689599000000000,
      "source_sequence": 100
    }
  ]
}
```

The index supports timestamp and sequence seeking. Implementations MAY build sidecar indexes for large archives. Sidecar indexes are not part of MRX v0 compatibility.

## Sequence Semantics

`source_sequence` stores the provider's sequence when available. It is not required to be contiguous across every provider or source.

For sources with contiguous sequences, readers and replay tools SHOULD detect:

- duplicate sequence values
- missing sequence gaps
- decreasing sequence values

Gap detection requires source-specific knowledge and is not a generic file-corruption error.

## Seeking

Readers can seek by:

- byte offset to a known frame boundary
- `frame_number`
- `receive_time_ns`
- `source_time_ns`
- `source_sequence`

Without an index, seeking is linear scan. With an index, seeking MAY use binary search over sorted index entries.

## Rollover

Writers SHOULD roll files by size, time or source partition. A new file MUST have a new `file_id`.

Rollover metadata SHOULD include:

- `previous_file_id`
- `next_file_id`
- `rollover_reason`
- `first_source_sequence`
- `last_source_sequence`

These fields live in header metadata or object-store metadata. They are optional in MRX v0.

## Truncation And Recovery

Readers MUST handle:

- missing footer
- incomplete final frame
- unsupported compression
- checksum mismatch
- invalid magic values
- frame length overflow

Recovery rule: return all complete frames before the first invalid or incomplete frame, then report the archive status.

Archive statuses:

- `CLEAN`: valid footer and all frames valid
- `UNCLEAN_RECOVERABLE`: no valid footer but at least one complete frame recovered
- `EMPTY_OR_UNREADABLE`: no complete frame recovered
- `CORRUPT`: invalid data before a recoverable frame boundary or checksum failure in a non-final complete frame

## Compatibility

Readers MUST reject files with an unsupported major version.

Readers SHOULD accept newer minor versions when:

- fixed fields remain compatible
- unknown flags are zero or ignorable
- unknown metadata keys can be preserved or ignored safely

Writers MUST set reserved flags to zero.

## Minimal Example

The smallest useful MRX file contains:

- file header with `NONE` compression support implied
- one frame containing original provider bytes
- footer

Example raw payload:

```json
{"price":42.5,"symbol":"AAPL"}
```

The payload recovered from the MRX frame MUST be byte-identical to the original UTF-8 bytes, including field order and spacing.

## Test Vectors

Normative test vectors live in [mrx-v0-test-vectors.json](mrx-v0-test-vectors.json).

Implementations SHOULD pass these behaviors before claiming MRX v0 compatibility:

- read a valid `NONE` compressed single-frame archive
- recover payload bytes exactly
- reject a payload checksum mismatch
- recover complete frames from a missing-footer archive
- report unsupported compression separately from corruption
