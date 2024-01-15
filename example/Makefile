BaseDir := $(CURDIR)/..
playbooks := bootstrap-os.yaml

.PHONY: build
build:
	go build -o $(BaseDir)/example -gcflags all=-N  github.com/kubesphere/kubekey/v4/cmd/kk

.PHONY: run-playbook
run-playbook: build
	@for pb in $(playbooks); do \
		$(BaseDir)/example/kk  run --work-dir=$(BaseDir)/example/test \
			--project-addr=git@github.com:littleBlackHouse/kse-installer.git \
			--project-branch=demo --inventory=$(BaseDir)/example/inventory.yaml \
			--config=$(BaseDir)/example/config.yaml \
			--debug playbooks/$$pb;\
    done

.PHONY: precheck
precheck: build
	$(BaseDir)/example/kk precheck --work-dir=$(BaseDir)/example/test \
	 --inventory=$(BaseDir)/example/inventory.yaml \
	 --config=$(BaseDir)/example/config.yaml

