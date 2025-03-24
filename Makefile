# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL:=help

#
# Go.
#
GO_VERSION ?= 1.23.3
GO_CONTAINER_IMAGE ?= docker.io/library/golang:$(GO_VERSION)
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn,direct
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
#export GO111MODULE=on

# This option is for running docker manifest command
#export DOCKER_CLI_EXPERIMENTAL := enabled

#
# Directories.
#
# Full directory of where the Makefile resides
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
#EXP_DIR := exp

TEST_DIR := test
TOOLS_DIR := hack/tools
#BIN_DIR := $(abspath $(TOOLS_DIR)/$(BIN_DIR))
E2E_FRAMEWORK_DIR := $(TEST_DIR)/framework
GO_INSTALL := ./hack/go_install.sh

# output
OUTPUT_DIR := $(abspath $(ROOT_DIR)/_output)
OUTPUT_BIN_DIR := $(OUTPUT_DIR)/bin
OUTPUT_TOOLS_DIR := $(OUTPUT_DIR)/tools
#ARTIFACTS ?= ${OUTPUT_DIR}/_artifacts

dirs := $(OUTPUT_DIR) $(OUTPUT_BIN_DIR) $(OUTPUT_TOOLS_DIR)

$(foreach dir, $(dirs), \
  $(if $(shell [ -d $(dir) ] && echo 1 || echo 0),, \
    $(shell mkdir -p $(dir)) \
  ) \
)

export PATH := $(abspath $(OUTPUT_BIN_DIR)):$(abspath $(OUTPUT_TOOLS_DIR)):$(PATH)

#
# Binaries.
#
# Note: Need to use abspath so we can invoke these from subdirectories
KUSTOMIZE_VER := v5.5.0
KUSTOMIZE_BIN := kustomize
KUSTOMIZE := $(abspath $(OUTPUT_TOOLS_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER))
KUSTOMIZE_PKG := sigs.k8s.io/kustomize/kustomize/v5

SETUP_ENVTEST_VER := v0.0.0-20240521074430-fbb7d370bebc
SETUP_ENVTEST_BIN := setup-envtest
SETUP_ENVTEST := $(abspath $(OUTPUT_TOOLS_DIR)/$(SETUP_ENVTEST_BIN)-$(SETUP_ENVTEST_VER))
SETUP_ENVTEST_PKG := sigs.k8s.io/controller-runtime/tools/setup-envtest

CONTROLLER_GEN_VER := main
CONTROLLER_GEN_BIN := controller-gen
CONTROLLER_GEN := $(abspath $(OUTPUT_TOOLS_DIR)/$(CONTROLLER_GEN_BIN)-$(CONTROLLER_GEN_VER))
CONTROLLER_GEN_PKG := sigs.k8s.io/controller-tools/cmd/controller-gen

GOTESTSUM_VER := v1.6.4
GOTESTSUM_BIN := gotestsum
GOTESTSUM := $(abspath $(OUTPUT_TOOLS_DIR)/$(GOTESTSUM_BIN)-$(GOTESTSUM_VER))
GOTESTSUM_PKG := gotest.tools/gotestsum

HADOLINT_VER := v2.10.0
HADOLINT_FAILURE_THRESHOLD = warning

GOLANGCI_LINT_VER := $(shell cat .github/workflows/golangci-lint.yaml | grep [[:space:]]version | sed 's/.*version: //')
GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(abspath $(OUTPUT_TOOLS_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER))
GOLANGCI_LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

GORELEASER_VER := $(shell cat .github/workflows/releaser.yaml | grep [[:space:]]version | sed 's/.*version: //')
GORELEASER_BIN := goreleaser
GORELEASER := $(abspath $(OUTPUT_TOOLS_DIR)/$(GORELEASER_BIN)-$(GORELEASER_VER))
GORELEASER_PKG := github.com/goreleaser/goreleaser/v2

#
# Docker.
#
DOCKERCMD ?= $(shell which docker)
DOCKER_BUILD_ENV = DOCKER_BUILDKIT=1
DOCKER_BUILD ?= $(DOCKER_BUILD_ENV) $(DOCKERCMD) buildx build
PLATFORM ?= linux/amd64,linux/arm64
DOCKER_OUT_TYPE ?= --push
DOCKER_PUSH ?= $(DOCKER_BUILD) --platform $(PLATFORM) $(DOCKER_OUT_TYPE)

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY ?= docker.io/kubespheredev

# capkk
CAPKK_CONTROLLER_IMG_NAME ?= capkk-controller-manager
CAPKK_CONTROLLER_IMG ?= $(REGISTRY)/$(CAPKK_CONTROLLER_IMG_NAME)
# controller-manager
KK_CONTROLLER_IMG_NAME ?= kk-controller-manager
KK_CONTROLLER_IMG ?= $(REGISTRY)/$(KK_CONTROLLER_IMG_NAME)
# executor
KK_EXECUTOR_IMG_NAME ?= kk-executor
KK_EXECUTOR_IMG ?= $(REGISTRY)/$(KK_EXECUTOR_IMG_NAME)

# It is set by Prow GIT_TAG, a git-based tag of the form vYYYYMMDD-hash, e.g., v20210120-v0.3.10-308-gc61521971

TAG ?= dev

# Set build time variables including version details
LDFLAGS := $(shell hack/version.sh)
# Set kk build tags
#BUILDTAGS = exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp
BUILDTAGS ?= builtin
#.PHONY: all
#all: test managers

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Generate / Manifests
## --------------------------------------:

##@ generate:

.PHONY: generate
generate: generate-go-deepcopy generate-manifests-kubekey generate-manifests-capkk generate-modules generate-goimports ## Run all generate-manifests-*, generate-go-deepcopy-* targets

.PHONY: generate-go-deepcopy
generate-go-deepcopy: $(CONTROLLER_GEN) ## Generate deepcopy object
	$(MAKE) clean-generated-deepcopy SRC_DIRS="./api/"
	@$(CONTROLLER_GEN) \
		object:headerFile=./hack/boilerplate.go.txt \
		paths=./api/...

.PHONY: generate-manifests-kubekey
generate-manifests-kubekey: $(CONTROLLER_GEN) clean-crds-kubekey ## Generate kubekey manifests e.g. CRD, RBAC etc.
	@$(CONTROLLER_GEN) \
		paths=./api/core/... \
		crd output:crd:dir=./config/kubekey/crds/
	@$(CONTROLLER_GEN) \
		paths=./pkg/controllers/core/... \
		rbac:roleName=kubekey output:rbac:dir=./config/kubekey/templates/

.PHONY: generate-manifests-capkk
generate-manifests-capkk: $(CONTROLLER_GEN) $(KUSTOMIZE) clean-crds-capkk ## Generate capkk manifests e.g. CRD, RBAC etc.
	@$(CONTROLLER_GEN) \
		paths=./api/capkk/... \
		crd \
		output:crd:dir=./config/capkk/crds/ 
	@$(CONTROLLER_GEN) \
		paths=./pkg/controllers/... \
		rbac:roleName=capkk output:rbac:dir=./config/capkk/rbac \
		webhook output:webhook:dir=./config/capkk/webhook
	@cp ./config/kubekey/crds/* ./config/capkk/crds/
	@cd config/capkk && $(KUSTOMIZE) edit set image capkk-controller-manager-image=$(CAPKK_CONTROLLER_IMG):$(TAG) kk-controller-manager-image=$(KK_CONTROLLER_IMG):$(TAG)
	@$(KUSTOMIZE) build config/capkk | \
		yq eval '.metadata |= select(.name == "default-capkk") *+ {"annotations": {"cert-manager.io/inject-ca-from": "capkk-system/capkk-serving-cert"}}' | \
		yq eval '.spec.template.spec.containers[] |= (select(.name == "controller-manager") | .env[] |= (select(.name == "EXECUTOR_IMAGE") | .value = "$(KK_EXECUTOR_IMG):$(TAG)"))' \
		> config/capkk/release/infrastructure-components.yaml

.PHONY: generate-modules
generate-modules: ## Run go mod tidy to ensure modules are up to date
	@cd api && go mod tidy
	@go mod tidy
.PHONY: generate-goimports
generate-goimports:  ## Format all import, `goimports` is required.
	@hack/update-goimports.sh

## --------------------------------------
## Lint / Verify
## --------------------------------------

##@ lint and verify:

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the codebase
	@$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS) 
	@cd $(TEST_DIR); $(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: verify-dockerfiles
verify-dockerfiles:
	@./hack/ci-lint-dockerfiles.sh $(HADOLINT_VER) $(HADOLINT_FAILURE_THRESHOLD)

ALL_VERIFY_CHECKS ?= modules gen goimports releaser

.PHONY: verify
verify: $(addprefix verify-,$(ALL_VERIFY_CHECKS))  ## Run all verify-* targets

.PHONY: verify-modules
verify-modules:  ## Verify go modules are up to date
	@if !(git diff --quiet HEAD -- go.sum go.mod $(TOOLS_DIR)/go.mod $(TOOLS_DIR)/go.sum $(TEST_DIR)/go.mod $(TEST_DIR)/go.sum); then \
		git diff; \
		echo "go module files are out of date"; exit 1; \
	fi
	@if (find . -name 'go.mod' | xargs -n1 grep -q -i 'k8s.io/client-go.*+incompatible'); then \
		find . -name "go.mod" -exec grep -i 'k8s.io/client-go.*+incompatible' {} \; -print; \
		echo "go module contains an incompatible client-go version"; exit 1; \
	fi

.PHONY: verify-gen
verify-gen:  ## Verify go generated files are up to date
	@if !(git diff --quiet HEAD); then \
		git diff; \
		echo "generated files are out of date, run make generate"; exit 1; \
	fi

.PHONY: verify-goimports
verify-goimports: ## Verify go imports
	@hack/verify-goimports.sh

.PHONY: verify-releaser
verify-releaser: $(GORELEASER) ## Verify goreleaser
	@$(GORELEASER) check

## --------------------------------------
## Binaries
## --------------------------------------

##@ build:

.PHONY: kk
kk: ## build kk binary
	@CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build -trimpath -tags "$(BUILDTAGS)" -ldflags "$(LDFLAGS)" -o $(OUTPUT_BIN_DIR)/kk cmd/kk/kubekey.go

.PHONY: kk-releaser 
kk-releaser: $(GORELEASER) ## build releaser in dist. it will show in https://github.com/kubesphere/kubekey/releases
	@LDFLAGS=$(bash ./hack/version.sh) $(GORELEASER) release --clean --skip validate --skip publish

.PHONY: docker-push ## build and push all images
docker-push: docker-push-kk-executor docker-push-kk-controller-manager docker-push-capkk-controller-manager

.PHONY: docker-push-kk-executor
docker-push-kk-executor: ## Build the docker image for kk-executor
	@$(DOCKER_PUSH) \
		--build-arg builder_image=$(GO_CONTAINER_IMAGE) \
		--build-arg goproxy=$(GOPROXY) \
		--build-arg ldflags="$(LDFLAGS)" --build-arg build_tags="" \
		-f build/kk/Dockerfile -t $(KK_EXECUTOR_IMG):$(TAG) .

.PHONY: docker-push-kk-controller-manager
docker-push-kk-controller-manager: ## Build the docker image for kk-controller-manager
	@$(DOCKER_PUSH) \
		--build-arg builder_image=$(GO_CONTAINER_IMAGE) \
		--build-arg goproxy=$(GOPROXY) \
		--build-arg ldflags="$(LDFLAGS)" --build-arg build_tags="builtin" \
		-f build/controller-manager/Dockerfile -t $(KK_CONTROLLER_IMG):$(TAG) .

.PHONY: docker-push-capkk-controller-manager
docker-push-capkk-controller-manager: ## Build the docker image for capkk-controller-manager
	@$(DOCKER_PUSH) \
		--build-arg builder_image=$(GO_CONTAINER_IMAGE) \
		--build-arg goproxy=$(GOPROXY) \
		--build-arg ldflags="$(LDFLAGS)" --build-arg build_tags="clusterapi" \
		-f build/controller-manager/Dockerfile -t $(CAPKK_CONTROLLER_IMG):$(TAG) .

## --------------------------------------
## Deployment
## --------------------------------------

##@ deployment

.PHONY: helm-package
helm-package: ## Helm-package.
	@helm package config/helm -d $(OUTPUT_DIR)

## --------------------------------------
## Testing
## --------------------------------------

##@ test:

.PHONY: test
test: $(SETUP_ENVTEST) ## Run unit and integration tests
	@KUBEBUILDER_ASSETS="$(KUBEBUILDER_ASSETS)" go test ./... $(TEST_ARGS)

.PHONY: test-verbose
test-verbose: ## Run unit and integration tests with verbose flag
	@$(MAKE) test TEST_ARGS="$(TEST_ARGS) -v"

## --------------------------------------
## Cleanup / Verification
## --------------------------------------

##@ clean:

.PHONY: clean
clean: clean-output clean-generated-deepcopy clean-crds-kubekey clean-crds-capkk ## Remove all generated files

.PHONY: clean-output
clean-output: ## Remove all generated binaries
	@rm -rf $(OUTPUT_DIR)

.PHONY: clean-crds-kubekey
clean-crds-kubekey: ## Remove the generated crds for kubekey
	@rm -rf ./config/kubekey/crds

.PHONY: clean-crds-capkk
clean-crds-capkk: ## Remove the generated crds for capkk
	@rm -rf ./config/capkk/crds

#.PHONY: clean-release-git
#clean-release-git: ## Restores the git files usually modified during a release
#	git restore ./*manager_image_patch.yaml ./*manager_pull_policy.yaml
#
#.PHONY: clean-generated-yaml
#clean-generated-yaml: ## Remove files generated by conversion-gen from the mentioned dirs. Example SRC_DIRS="./api/v1beta1"
#	(IFS=','; for i in $(SRC_DIRS); do find $$i -type f -name '*.yaml' -exec rm -f {} \;; done)
#
.PHONY: clean-generated-deepcopy
clean-generated-deepcopy: ## Remove files generated by conversion-gen from the mentioned dirs. Example SRC_DIRS="./api/v1beta1"
	@(IFS=','; for i in $(SRC_DIRS); do find $$i -type f -name 'zz_generated.deepcopy*' -exec rm -f {} \;; done)

## --------------------------------------
## Hack / Tools
## --------------------------------------

##@ hack/tools:

$(CONTROLLER_GEN): # Build controller-gen into tools folder.
	@if [ ! -f $(CONTROLLER_GEN) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(CONTROLLER_GEN_PKG) $(CONTROLLER_GEN_BIN) $(CONTROLLER_GEN_VER); \
	fi

$(GOTESTSUM): # Build gotestsum into tools folder.
	@if [ ! -f $(GOTESTSUM) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(GOTESTSUM_PKG) $(GOTESTSUM_BIN) $(GOTESTSUM_VER); \
	fi

$(KUSTOMIZE): # Build kustomize into tools folder.
	@if [ ! -f $(KUSTOMIZE) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(KUSTOMIZE_PKG) $(KUSTOMIZE_BIN) $(KUSTOMIZE_VER); \
	fi

$(SETUP_ENVTEST): # Build setup-envtest into tools folder.
	if [ ! -f $(SETUP_ENVTEST) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(SETUP_ENVTEST_PKG) $(SETUP_ENVTEST_BIN) $(SETUP_ENVTEST_VER); \
	fi

$(GOLANGCI_LINT): # Build golangci-lint into tools folder.
	@if [ ! -f $(GOLANGCI_LINT) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(GOLANGCI_LINT_PKG) $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER); \
	fi

$(GORELEASER): # Build goreleaser into tools folder.
	@if [ ! -f $(GORELEASER) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(GORELEASER_PKG) $(GORELEASER_BIN) $(GORELEASER_VER); \
	fi
