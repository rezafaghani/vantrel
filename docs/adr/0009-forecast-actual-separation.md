# ADR 0009: Forecast And Actual Separation

Status: Accepted

## Context

Forecasts and realized observations have different semantics and are compared later for error, bias, directional accuracy and PnL impact.

## Decision

Model forecast observations separately from actual observations. Link forecast series to actual series through Catalog relationships.

## Consequences

- Forecast points reference ForecastRun metadata.
- Actual observations preserve event/delivery time and revision where applicable.
- Ambiguous generic tables for both categories are avoided.
