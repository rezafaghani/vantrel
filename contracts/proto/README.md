# Canonical Market Contracts

Vantrel canonical contracts use Protobuf for language-neutral market events.

Initial package:

- `vantrel.market.v1`

Current contracts cover:

- event envelope
- money and quantity value types
- trade
- quote
- order-book delta
- order-book snapshot
- forecast observation
- actual observation
- quality metadata

Rules:

- Raw provider payloads stay in MRX; these contracts are normalized canonical events.
- Series meaning belongs in the Catalog, not in observation messages.
- Forecast and actual observations remain separate messages.
- Field numbers `100` to `199` are reserved in top-level market events for future compatibility escape hatches.

Generated code is intentionally not committed until the repository has a pinned Protobuf toolchain.
