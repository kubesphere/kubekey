# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL:=help

#
# Go.
#
GO_VERSION ?= 1.18.3
GO_CONTAINER_IMAGE ?= docker.io/library/golang:$(GO_VERSION)

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn,direct
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

# This option is for running docker manifest command
export DOCKER_CLI_EXPERIMENTAL := enabled

#
# Directories.
#
# Full directory of where the Makefile resides
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
EXP_DIR := exp
BIN_DIR := bin
TEST_DIR := test
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/$(BIN_DIR))
E2E_FRAMEWORK_DIR := $(TEST_DIR)/framework
GO_INSTALL := ./scripts/go_install.sh

TAG ?= dev
ARCH ?= $(shell go env GOARCH)
ALL_ARCH = amd64 arm arm64 ppc64le s390x

#
# Binaries.
#
# Note: Need to use abspath so we can invoke these from subdirectories
KUSTOMIZE_VER := v4.5.2
KUSTOMIZE_BIN := kustomize
KUSTOMIZE := $(abspath $(TOOLS_BIN_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER))
KUSTOMIZE_PKG := sigs.k8s.io/kustomize/kustomize/v4

SETUP_ENVTEST_VER := v0.0.0-20211110210527-619e6b92dab9
SETUP_ENVTEST_BIN := setup-envtest
SETUP_ENVTEST := $(abspath $(TOOLS_BIN_DIR)/$(SETUP_ENVTEST_BIN)-$(SETUP_ENVTEST_VER))
SETUP_ENVTEST_PKG := sigs.k8s.io/controller-runtime/tools/setup-envtest

CONTROLLER_GEN_VER := v0.9.1
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(abspath $(TOOLS_BIN_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER))
CONTROLLER_GEN_PKG := sigs.k8s.io/controller-tools/cmd/controller-gen

GOTESTSUM_VER := v1.6.4
GOTESTSUM_BIN := gotestsum
GOTESTSUM := $(abspath $(TOOLS_BIN_DIR)/$(GOTESTSUM_BIN)-$(GOTESTSUM_VER))
GOTESTSUM_PKG := gotest.tools/gotestsum

HADOLINT_VER := v2.10.0
HADOLINT_FAILURE_THRESHOLD = warning

GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN))

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY ?= docker.io/kubespheredev
PROD_REGISTRY ?= docker.io/kubesphere

# capkk
CAPKK_IMAGE_NAME ?= capkk-manager
CAPKK_CONTROLLER_IMG ?= $(REGISTRY)/$(CAPKK_IMAGE_NAME)

# It is set by Prow GIT_TAG, a git-based tag of the form vYYYYMMDD-hash, e.g., v20210120-v0.3.10-308-gc61521971

TAG ?= dev
ARCH ?= $(shell go env GOARCH)
ALL_ARCH = amd64 arm arm64 ppc64le s390x

# Set build time variables including version details
LDFLAGS := $(shell hack/version.sh)

.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Generate / Manifests
## --------------------------------------

##@ generate:

.PHONY: manifests
manifests: $(CONTROLLER_GEN) ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: $(CONTROLLER_GEN) ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

## --------------------------------------
## Lint / Verify
## --------------------------------------

##@ lint and verify:

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the codebase
	$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)
	cd $(TEST_DIR); $(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)
	cd $(TOOLS_DIR); $(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)
	./scripts/ci-lint-dockerfiles.sh $(HADOLINT_VER) $(HADOLINT_FAILURE_THRESHOLD)

.PHONY: lint-dockerfiles
lint-dockerfiles:
	./scripts/ci-lint-dockerfiles.sh $(HADOLINT_VER) $(HADOLINT_FAILURE_THRESHOLD)

.PHONY: verify
verify: $(addprefix verify-,$(ALL_VERIFY_CHECKS)) lint-dockerfiles ## Run all verify-* targets

.PHONY: verify-modules
verify-modules: generate-modules  ## Verify go modules are up to date
	@if !(git diff --quiet HEAD -- go.sum go.mod $(TOOLS_DIR)/go.mod $(TOOLS_DIR)/go.sum $(TEST_DIR)/go.mod $(TEST_DIR)/go.sum); then \
		git diff; \
		echo "go module files are out of date"; exit 1; \
	fi
	@if (find . -name 'go.mod' | xargs -n1 grep -q -i 'k8s.io/client-go.*+incompatible'); then \
		find . -name "go.mod" -exec grep -i 'k8s.io/client-go.*+incompatible' {} \; -print; \
		echo "go module contains an incompatible client-go version"; exit 1; \
	fi

.PHONY: verify-gen
verify-gen: generate  ## Verify go generated files are up to date
	@if !(git diff --quiet HEAD); then \
		git diff; \
		echo "generated files are out of date, run make generate"; exit 1; \
	fi

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build:  ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle $(BUNDLE_GEN_FLAGS)
	operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.19.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

## --------------------------------------
## Binaries
## --------------------------------------

##@ build:

ALL_MANAGERS = capkk

.PHONY: managers
managers: $(addprefix manager-,$(ALL_MANAGERS)) ## Run all manager-* targets

.PHONY: manager-capkk
manager-capkk: ## Build the capkk manager binary into the ./bin folder
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/manager kubesphere.io/kubekey/exp/cluster-api-provider-kubekey

.PHONY: docker-pull-prerequisites
docker-pull-prerequisites:
	docker pull docker.io/docker/dockerfile:1.4
	docker pull $(GO_CONTAINER_IMAGE)

.PHONY: docker-build-all
docker-build-all: $(addprefix docker-build-,$(ALL_ARCH)) ## Build docker images for all architectures

docker-build-%:
	$(MAKE) ARCH=$* docker-build

ALL_DOCKER_BUILD = capkk

.PHONY: docker-build-capkk
docker-build-capkk: ## Build the docker image for capkk
	DOCKER_BUILDKIT=1 docker build --build-arg builder_image=$(GO_CONTAINER_IMAGE) --build-arg goproxy=$(GOPROXY) --build-arg ARCH=$(ARCH) --build-arg ldflags="$(LDFLAGS)" . -t $(CAPKK_CONTROLLER_IMG)-$(ARCH):$(TAG)
	$(MAKE) set-manifest-image MANIFEST_IMG=$(CAPKK_CONTROLLER_IMG)-$(ARCH) MANIFEST_TAG=$(TAG)
	$(MAKE) set-manifest-pull-policy TARGET_RESOURCE="./config/default/manager_pull_policy.yaml"

.PHONY: docker-build-e2e
docker-build-e2e: ## Build the docker image for capkk
	DOCKER_BUILDKIT=1 docker build --build-arg builder_image=$(GO_CONTAINER_IMAGE) --build-arg goproxy=$(GOPROXY) --build-arg ARCH=$(ARCH) --build-arg ldflags="$(LDFLAGS)" . -t "$(CAPKK_CONTROLLER_IMG):e2e"
	$(MAKE) set-manifest-image MANIFEST_IMG=$(CAPKK_CONTROLLER_IMG) MANIFEST_TAG="e2e" TARGET_RESOURCE="./config/default/manager_image_patch.yaml"
	$(MAKE) set-manifest-pull-policy PULL_POLICY=IfNotPresent TARGET_RESOURCE="./config/default/manager_pull_policy.yaml"

## --------------------------------------
## Testing
## --------------------------------------

##@ test:

ARTIFACTS ?= ${ROOT_DIR}/_artifacts

ifeq ($(shell go env GOOS),darwin) # Use the darwin/amd64 binary until an arm64 version is available
	KUBEBUILDER_ASSETS ?= $(shell $(SETUP_ENVTEST) use --use-env -p path --arch amd64 $(KUBEBUILDER_ENVTEST_KUBERNETES_VERSION))
else
	KUBEBUILDER_ASSETS ?= $(shell $(SETUP_ENVTEST) use --use-env -p path $(KUBEBUILDER_ENVTEST_KUBERNETES_VERSION))
endif

.PHONY: test
test: $(SETUP_ENVTEST) ## Run unit and integration tests
	KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test ./... $(TEST_ARGS)

.PHONY: test-verbose
test-verbose: ## Run unit and integration tests with verbose flag
	$(MAKE) test TEST_ARGS="$(TEST_ARGS) -v"

.PHONY: test-junit
test-junit: $(SETUP_ENVTEST) $(GOTESTSUM) ## Run unit and integration tests and generate a junit report
	set +o errexit; (KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test -json ./... $(TEST_ARGS); echo $$? > $(ARTIFACTS)/junit.exitcode) | tee $(ARTIFACTS)/junit.stdout
	$(GOTESTSUM) --junitfile $(ARTIFACTS)/junit.xml --raw-command cat $(ARTIFACTS)/junit.stdout
	exit $$(cat $(ARTIFACTS)/junit.exitcode)

.PHONY: test-cover
test-cover: ## Run unit and integration tests and generate a coverage report
	$(MAKE) test TEST_ARGS="$(TEST_ARGS) -coverprofile=out/coverage.out"
	go tool cover -func=out/coverage.out -o out/coverage.txt
	go tool cover -html=out/coverage.out -o out/coverage.html

.PHONY: test-e2e
test-e2e: ## Run e2e tests
	$(MAKE) -C $(TEST_DIR)/e2e run

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: set-manifest-pull-policy
set-manifest-pull-policy:
	$(info Updating kustomize pull policy file for manager resources)
	sed -i'' -e 's@imagePullPolicy: .*@imagePullPolicy: '"$(PULL_POLICY)"'@' $(TARGET_RESOURCE)

.PHONY: set-manifest-image
set-manifest-image:
	$(info Updating kustomize image patch file for manager resource)
	sed -i'' -e 's@image: .*@image: '"${MANIFEST_IMG}:$(MANIFEST_TAG)"'@' $(TARGET_RESOURCE)

## --------------------------------------
## Hack / Tools
## --------------------------------------

##@ hack/tools:

.PHONY: $(CONTROLLER_GEN_BIN)
$(CONTROLLER_GEN_BIN): $(CONTROLLER_GEN) ## Build a local copy of controller-gen.

.PHONY: $(GOTESTSUM_BIN)
$(GOTESTSUM_BIN): $(GOTESTSUM) ## Build a local copy of gotestsum.

.PHONY: $(KUSTOMIZE_BIN)
$(KUSTOMIZE_BIN): $(KUSTOMIZE) ## Build a local copy of kustomize.

.PHONY: $(SETUP_ENVTEST_BIN)
$(SETUP_ENVTEST_BIN): $(SETUP_ENVTEST) ## Build a local copy of setup-envtest.

.PHONY: $(GOLANGCI_LINT_BIN)
$(GOLANGCI_LINT_BIN): $(GOLANGCI_LINT) ## Build a local copy of golangci-lint

$(CONTROLLER_GEN): # Build controller-gen from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(CONTROLLER_GEN_PKG) $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER)

$(GOTESTSUM): # Build gotestsum from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GOTESTSUM_PKG) $(GOTESTSUM_BIN) $(GOTESTSUM_VER)

$(KUSTOMIZE): # Build kustomize from tools folder.
	CGO_ENABLED=0 GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(KUSTOMIZE_PKG) $(KUSTOMIZE_BIN) $(KUSTOMIZE_VER)

$(SETUP_ENVTEST): # Build setup-envtest from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(SETUP_ENVTEST_PKG) $(SETUP_ENVTEST_BIN) $(SETUP_ENVTEST_VER)

$(GOLANGCI_LINT): ../../.github/workflows/golangci-lint.yml # Download golangci-lint using hack script into tools folder.
	hack/ensure-golangci-lint.sh \
		-b $(TOOLS_BIN_DIR) \
		$(shell cat .github/workflows/golangci-lint.yml | grep [[:space:]]version | sed 's/.*version: //')