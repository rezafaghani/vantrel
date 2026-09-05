KIND_CLUSTER ?= vantrel
KIND_CONFIG ?= deployments/kind/vantrel-kind.yaml
HELM_RELEASE ?= vantrel-platform
HELM_CHART ?= charts/vantrel-platform
HELM_NAMESPACE ?= default
GOCACHE ?= /tmp/vantrel-go-cache

.PHONY: k8s-check-tools
k8s-check-tools:
	command -v kubectl
	command -v kind
	command -v helm

.PHONY: kind-create
kind-create: k8s-check-tools
	kind create cluster --name $(KIND_CLUSTER) --config $(KIND_CONFIG)

.PHONY: kind-delete
kind-delete:
	kind delete cluster --name $(KIND_CLUSTER)

.PHONY: helm-lint
helm-lint:
	helm lint $(HELM_CHART)

.PHONY: helm-template
helm-template:
	helm template $(HELM_RELEASE) $(HELM_CHART)

.PHONY: helm-install
helm-install: k8s-check-tools
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) --namespace $(HELM_NAMESPACE)

.PHONY: k8s-status
k8s-status:
	kubectl get namespaces -l app.kubernetes.io/part-of=vantrel

.PHONY: proto-check
proto-check:
	python3 tools/check_proto_compat.py
	python3 -m unittest tools/check_proto_compat_test.py

.PHONY: go-test
go-test:
	cd libs/go/mrx && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/catalog && GOCACHE=$(GOCACHE) go test ./...

.PHONY: go-vet
go-vet:
	cd libs/go/mrx && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/catalog && GOCACHE=$(GOCACHE) go vet ./...
