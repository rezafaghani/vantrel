# ADR 0001: Kubernetes-first

Status: Accepted

## Context

Vantrel is intended to run as independently deployable services with strong operational contracts.

## Decision

Use Kubernetes as the primary runtime model. Local development will target kind or k3d with Helm conventions.

The local foundation is documented in [Local Kubernetes Development](../kubernetes/local-development.md).

## Consequences

- Services must define health, readiness, configuration, persistence and graceful shutdown.
- Docker Compose is not the primary development runtime.
- The initial repository does not install a cluster or platform stack.
