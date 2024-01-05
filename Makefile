# Image URL to use all building/pushing image targets
OB_REPO ?= ghcr.io/oneblock-ai
OB_TAG ?= latest
OB_IMG ?= ${OB_REPO}/oneblock:${OB_TAG}
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.28.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate Generate CRD manifest YAMLs.
	bash ./hack/generate-manifest

.PHONY: generate
generate: ## Generate code containing controller factories and manifest in generate.go
	go generate

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt lint vet manifests generate envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.54.2
golangci-lint:
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: ## Build oneblock locally using goreleaser.
	goreleaser release --snapshot --clean

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
.PHONY: buildx
buildx: ## Build and push docker image for the oneblock for cross-platform support
	OB_BUILD_VERSION = $(shell cat dist/metadata.json | jq -r .version)
	IMAGES := $(shell docker images --format "{{.Repository}}:{{.Tag}}" | grep $(OB_BUILD_VERSION))
	for i in $(IMAGES); do \
		echo $$i; \
		$(CONTAINER_TOOL) push $$i; \
	done
	$(CONTAINER_TOOL) manifest create $(OB_IMG) $(IMAGES)
	$(CONTAINER_TOOL) manifest push $(OB_IMG)


##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install-crds
install-crds: manifests ## Install CRDs into the K8s cluster.
	$(HELM) upgrade --install --create-namespace -n oneblock-system oneblock-crd deploy/charts/oneblock-crd --reuse-values

.PHONY: uninstall-crd
uninstall-crd: manifests ## Uninstall CRDs from the K8s cluster.
	$(HELM) uninstall -n oneblock-system oneblock

.PHONY: install-ob
install-ob: ## Deploy api-server to the K8s cluster.
	$(HELM) upgrade --install --create-namespace -n oneblock-system oneblock deploy/charts/oneblock --reuse-values

.PHONY: uninstall-ob
uninstall-ob: ## Undeploy api-server from the K8s cluster.
	$(HELM) uninstall -n oneblock-system oneblock

.PHONY: run-apiserver
run-apiserver: ## Undeploy api-server from the K8s cluster.
	go run main.go api-server --kubeconfig="~/.kube/config.yaml"

##@ Helm chart
.PHONY: helm-dep-update
helm-dep-update: ## Update oneblock helm dependency chart
	$(HELM) dep update ./deploy/charts/oneblock
.PHONY: helm-lint
helm-lint: ## Run Helm lint against oneblock chart
	$(HELM) lint ./deploy/charts/oneblock

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
GLOBALBIN ?= /usr/local/bin

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
HELM ?= $(GLOBALBIN)/helm

## Tool Versions
KUSTOMIZE_VERSION ?= v5.2.1
CONTROLLER_TOOLS_VERSION ?= v0.13.0
HELM_VERSION ?= v3.13.1

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed first.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: helm
helm: $(HELM) ## Download Helm locally if necessary.
$(HELM): $(GLOBALBIN)
	test -s $(GLOBALBIN)/helm && $(GLOBALBIN)/helm version --short | grep -q $(HELM_VERSION) || \
	curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
