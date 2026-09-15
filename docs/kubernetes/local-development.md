# Local Kubernetes Development

Vantrel is Kubernetes-first. Docker Compose is not the primary development runtime.

Task 5 adds only the local foundation:

- a kind cluster config
- a minimal Helm chart for Vantrel namespaces
- Makefile targets for local operations

It does not install Kafka, databases, observability, Keycloak, APISIX, OPA or application services.

## Prerequisites

- `kubectl`
- `kind`
- `helm`

## Commands

Create a local cluster:

```sh
make kind-create
```

Render the Helm chart:

```sh
make helm-template
```

Install the namespace foundation:

```sh
make helm-install
```

The foundation chart is installed from the existing `default` namespace and creates the Vantrel namespaces it owns.

Check Vantrel namespaces:

```sh
make k8s-status
```

Build the local pipeline image, load it into kind, and install the app with QuestDB:

```sh
make k8s-app-install
```

Expose the dashboard:

```sh
make k8s-app-port-forward
```

Open `http://localhost:8080`.

The dashboard exposes the current QuestDB status, observation counts, latest rows and a synthetic ingestion trigger.

Delete the local cluster:

```sh
make kind-delete
```

## Namespace Convention

Initial namespaces:

- `vantrel-system`
- `vantrel-data`
- `vantrel-trading`
- `vantrel-observability`

Future service charts should declare readiness, liveness, resource requests, resource limits, configuration, secret references, persistence needs and graceful shutdown behavior.
