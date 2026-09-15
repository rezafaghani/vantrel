KIND_CLUSTER ?= vantrel
KIND_CONFIG ?= deployments/kind/vantrel-kind.yaml
HELM_RELEASE ?= vantrel-platform
HELM_CHART ?= charts/vantrel-platform
HELM_NAMESPACE ?= default
GOCACHE ?= /tmp/vantrel-go-cache
QUESTDB_IMAGE ?= docker.io/questdb/questdb:10.0.1
QUESTDB_CONTAINER ?= vantrel-questdb
LOCAL_PIPELINE_IMAGE ?= localhost/vantrel-local-pipeline:dev
APP_ADDR ?= :8080

.PHONY: k8s-check-tools
k8s-check-tools:
	command -v kubectl
	command -v kind

.PHONY: helm-check-tools
helm-check-tools:
	command -v helm

.PHONY: kind-create
kind-create: k8s-check-tools
	kind create cluster --name $(KIND_CLUSTER) --config $(KIND_CONFIG)

.PHONY: kind-delete
kind-delete:
	kind delete cluster --name $(KIND_CLUSTER)

.PHONY: helm-lint
helm-lint: helm-check-tools
	helm lint $(HELM_CHART)

.PHONY: helm-template
helm-template: helm-check-tools
	helm template $(HELM_RELEASE) $(HELM_CHART)

.PHONY: helm-install
helm-install: k8s-check-tools helm-check-tools
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) --namespace $(HELM_NAMESPACE)

.PHONY: k8s-status
k8s-status:
	kubectl get namespaces -l app.kubernetes.io/part-of=vantrel

.PHONY: proto-check
proto-check:
	python3 tools/check_proto_compat.py
	python3 -m unittest tools/check_proto_compat_test.py

.PHONY: kafka-check
kafka-check:
	python3 tools/check_kafka_topics.py
	python3 -m unittest tools/check_kafka_topics_test.py

.PHONY: questdb-schema-check
questdb-schema-check:
	python3 tools/check_questdb_schema.py
	python3 -m unittest tools/check_questdb_schema_test.py

.PHONY: go-test
go-test:
	cd libs/go/mrx && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/catalog && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/ingestion && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/synthetic && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/rawarchive && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/validation && GOCACHE=$(GOCACHE) go test ./...
	cd libs/go/marketstore && GOCACHE=$(GOCACHE) go test ./...
	cd services/catalog-api && GOCACHE=$(GOCACHE) go test ./...
	cd services/local-pipeline && GOCACHE=$(GOCACHE) go test ./...

.PHONY: go-vet
go-vet:
	cd libs/go/mrx && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/catalog && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/ingestion && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/synthetic && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/rawarchive && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/validation && GOCACHE=$(GOCACHE) go vet ./...
	cd libs/go/marketstore && GOCACHE=$(GOCACHE) go vet ./...
	cd services/catalog-api && GOCACHE=$(GOCACHE) go vet ./...
	cd services/local-pipeline && GOCACHE=$(GOCACHE) go vet ./...

.PHONY: run-local-pipeline
run-local-pipeline:
	cd services/local-pipeline && GOCACHE=$(GOCACHE) go run . -steps 3

.PHONY: questdb-up
questdb-up:
	mkdir -p .vantrel/questdb
	podman run -d --replace --name $(QUESTDB_CONTAINER) -p 9000:9000 -p 9003:9003 -v $(PWD)/.vantrel/questdb:/var/lib/questdb:Z $(QUESTDB_IMAGE)

.PHONY: questdb-down
questdb-down:
	podman stop $(QUESTDB_CONTAINER)

.PHONY: run-app
run-app:
	cd services/local-pipeline && GOCACHE=$(GOCACHE) go run . -serve -addr $(APP_ADDR)

.PHONY: app-image-build
app-image-build:
	podman build -f services/local-pipeline/Dockerfile -t $(LOCAL_PIPELINE_IMAGE) .

.PHONY: kind-load-app-image
kind-load-app-image: app-image-build
	mkdir -p .vantrel
	podman save $(LOCAL_PIPELINE_IMAGE) -o .vantrel/local-pipeline-image.tar.tmp
	mv -f .vantrel/local-pipeline-image.tar.tmp .vantrel/local-pipeline-image.tar
	kind load image-archive .vantrel/local-pipeline-image.tar --name $(KIND_CLUSTER)

.PHONY: k8s-app-install
k8s-app-install: kind-load-app-image
	kubectl apply -f deployments/kubernetes/local-app.yaml
	kubectl -n vantrel-data rollout restart deploy/local-pipeline
	kubectl -n vantrel-data rollout status deploy/questdb --timeout=180s
	kubectl -n vantrel-data rollout status deploy/local-pipeline --timeout=180s

.PHONY: k8s-app-port-forward
k8s-app-port-forward:
	kubectl -n vantrel-data port-forward svc/local-pipeline 8080:8080
