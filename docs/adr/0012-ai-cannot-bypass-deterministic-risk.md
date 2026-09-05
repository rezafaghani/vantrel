# ADR 0012: AI Cannot Bypass Deterministic Risk

Status: Accepted

## Context

AI may be useful for analysis, forecasting and recommendations, but trading controls must remain deterministic and enforceable.

## Decision

AI may propose actions but cannot bypass authentication, authorization, risk limits, portfolio constraints or execution policy.

## Consequences

- Execution flows require deterministic validation and risk checks.
- AI proposals are inputs to controls, not replacements for controls.
- Audit trails must preserve proposal, validation and decision context.
