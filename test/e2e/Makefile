# Copyright 2022 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.DEFAULT_GOAL:=help

GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn,direct
endif
export GOPROXY

REPO_ROOT := $(shell git rev-parse --show-toplevel)

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Binaries
## --------------------------------------

TOOLS_DIR := $(REPO_ROOT)/hack/tools
BIN_DIR := bin
TOOLS_BIN_DIR := $(TOOLS_DIR)/$(BIN_DIR)
GINGKO_VER := v1.16.5
GINKGO_BIN := ginkgo
GINKGO := $(abspath $(TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINGKO_VER))
GINKGO_PKG := github.com/onsi/ginkgo/ginkgo
GO_INSTALL := $(REPO_ROOT)/scripts/go_install.sh
KUSTOMIZE_BIN := kustomize
KUSTOMIZE_VER := $(shell grep "^KUSTOMIZE_VER" $(REPO_ROOT)/Makefile | awk '{print $$NF}')
KUSTOMIZE := $(abspath $(TOOLS_BIN_DIR)/$(KUSTOMIZE_BIN)-$(KUSTOMIZE_VER))
KUSTOMIZE_PKG := sigs.k8s.io/kustomize/kustomize/v4
KIND := $(abspath $(TOOLS_BIN_DIR)/$(KIND_BIN))
KIND_BIN := kind
KIND_PKG := sigs.k8s.io/kind
KIND_VER := v0.14.0

$(GINKGO): # Build ginkgo from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GINKGO_PKG) $(GINKGO_BIN) $(GINGKO_VER)

$(KUSTOMIZE): # Build kustomize from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(KUSTOMIZE_PKG) $(KUSTOMIZE_BIN) $(KUSTOMIZE_VER)

$(KIND):
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(KIND_PKG) $(KIND_BIN) $(KIND_VER)

.PHONY: $(GINKGO_BIN)
$(GINKGO_BIN): $(GINKGO) ## Build a local copy of ginkgo

.PHONY: $(KUSTOMIZE_BIN)
$(KUSTOMIZE_BIN): $(KUSTOMIZE) ## Build a local copy of kustomize

.PHONY: $(KIND_BIN)
$(KIND_BIN): $(KIND) ## Build a local copy of KIND

## --------------------------------------
## Templates
## --------------------------------------

KUBEKEY_TEMPLATES := $(REPO_ROOT)/test/e2e/data/infrastructure-kubekey
KUBEKEY_K3S_TEMPLATES := $(REPO_ROOT)/test/e2e/data/k3s

.PHONY: cluster-templates
cluster-templates: $(KUSTOMIZE) cluster-templates-v1beta1 ## Generate cluster templates for all versions

.PHONY: cluster-templates-k3s
cluster-templates-k3s: $(KUSTOMIZE) cluster-templates-k3s-v1beta1

cluster-templates-v1beta1: $(KUSTOMIZE) ## Generate cluster templates for v1beta1
	$(KUSTOMIZE) build $(KUBEKEY_TEMPLATES)/v1beta1/cluster-template --load-restrictor LoadRestrictionsNone > $(KUBEKEY_TEMPLATES)/v1beta1/cluster-template.yaml

cluster-templates-k3s-v1beta1: $(KUSTOMIZE)
	$(KUSTOMIZE) build $(KUBEKEY_K3S_TEMPLATES)/v1beta1/cluster-template --load-restrictor LoadRestrictionsNone > $(KUBEKEY_K3S_TEMPLATES)/v1beta1/cluster-template.yaml

## --------------------------------------
## Testing
## --------------------------------------

GINKGO_FOCUS ?=
GINKGO_SKIP ?=
GINKGO_NODES ?= 1
E2E_CONF_FILE ?= ${REPO_ROOT}/test/e2e/config/e2e_conf.yaml
E2E_K3S_CONF_FILE ?= ${REPO_ROOT}/test/e2e/config/e2e_k3s_conf.yaml
ARTIFACTS ?= ${REPO_ROOT}/_artifacts
ARTIFACTS_K3S ?= ${REPO_ROOT}/_artifacts_k3s
SKIP_RESOURCE_CLEANUP ?= false
USE_EXISTING_CLUSTER ?= false
GINKGO_NOCOLOR ?= false

# to set multiple ginkgo skip flags, if any
ifneq ($(strip $(GINKGO_SKIP)),)
_SKIP_ARGS := $(foreach arg,$(strip $(GINKGO_SKIP)),-skip="$(arg)")
endif

.PHONY: run
run: $(GINKGO) $(KIND) cluster-templates ## Run the end-to-end tests
	$(GINKGO) -v -trace -tags=e2e -focus="$(GINKGO_FOCUS)" $(_SKIP_ARGS) -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) ./suites/capkk/... -- \
	    -e2e.artifacts-folder="$(ARTIFACTS)" \
	    -e2e.config="$(E2E_CONF_FILE)" \
	    -e2e.skip-resource-cleanup=$(SKIP_RESOURCE_CLEANUP) -e2e.use-existing-cluster=$(USE_EXISTING_CLUSTER)

.PHONY: run-k3s
run-k3s: $(GINKGO) $(KIND) cluster-templates-k3s ## Run the end-to-end tests
	$(GINKGO) -v -trace -tags=e2e -focus="$(GINKGO_FOCUS)" $(_SKIP_ARGS) -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) ./suites/k3s/... -- \
	    -e2e.artifacts-folder="$(ARTIFACTS_K3S)" \
	    -e2e.config="$(E2E_K3S_CONF_FILE)" \
	    -e2e.skip-resource-cleanup=$(SKIP_RESOURCE_CLEANUP) -e2e.use-existing-cluster=$(USE_EXISTING_CLUSTER)