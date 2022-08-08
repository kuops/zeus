# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

ifndef ignore-not-found
  ignore-not-found = false
endif

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
SWAG = $(shell pwd)/bin/swag
.PHONY: get-controller-gen
get-controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

.PHONY: crd-gen
crd-gen: get-controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="$(PROJECT_DIR)/pkg/kubernetes/apis/..." output:crd:artifacts:config=deploy/crds/

.PHONY: update-codegen
update-codegen:
	bash ./hack/update-codegen.sh

.PHONY: get-swag
get-swag: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(SWAG),github.com/swaggo/swag/cmd/swag@latest)

.PHONY: update-swagger
update-swagger: get-swag
	swag init -g ./cmd/main.go

.PHONY: install
install: crd-gen ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f deploy/crds
.PHONY: uninstall
uninstall: crd-gen ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f deploy/crds

GO_FILES := $(shell find . -type f -name '*.go' -not -path './vendor/*' -print)
.PHONY: fmt
fmt:
	gofmt -s -w ${GO_FILES}
.PHONY: update
update: update-swagger update-codegen crd-gen
run:
	go run $(PROJECT_DIR)/cmd/main.go
