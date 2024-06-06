BaseDir?=$(CURDIR)
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: build
build:
	GOOS=$(GOOS);GOARCH=$(GOARCH) go build -o $(BaseDir)/kk-alpha -tags builtin -gcflags all=-N  github.com/kubesphere/kubekey/v4/cmd/kk

.PHONY: run-playbook
run-playbook: build
	@for pb in $(playbooks); do \
		$(BaseDir)/kk-alpha  run --work-dir=$(BaseDir)/test \
			--project-addr=git@github.com:littleBlackHouse/kse-installer.git \
			--project-branch=demo  \
			--debug playbooks/$$pb;\
    done

.PHONY: precheck
precheck: build
	$(BaseDir)/kk-alpha precheck --work-dir=$(BaseDir)/example/test \
	 --inventory=$(BaseDir)/example/inventory.yaml --debug

.PHONY: create-cluster
create-cluster: build
	$(BaseDir)/kk-alpha create cluster --work-dir=$(BaseDir)/example/test \
	 --inventory=$(BaseDir)/example/inventory.yaml \
	--debug

.PHONY: run-plugin
run-local: build
	@$(BaseDir)/kk-alpha  run --work-dir=$(BaseDir)/example/test \
     	--project-addr=$(BaseDir)/plugins \
     	--inventory=$(BaseDir)/example/inventory.yaml \
     	--set sonobuoy.plugins.kube_bench.enabled=true \
     	playbooks/sonobuoy.yaml
