.PHONY: create-role
create-role: ## create a role necessary file in roles
	@echo "Creating role $(role) in ${base} ..."
	@mkdir -p ${base}/roles/$(role)/tasks
	@echo "---" > ${base}/roles/$(role)/tasks/main.yaml
	@mkdir -p ${base}/roles/$(role)/defaults
	@echo "" > ${base}/roles/$(role)/defaults/main.yaml
ifeq ($(VARIABLE_NAME),"full")
	@mkdir -p ${base}/roles/$(role)/handlers
	@mkdir -p ${base}/roles/$(role)/templates
	@mkdir -p ${base}/roles/$(role)/files
	@mkdir -p ${base}/roles/$(role)/vars
	@mkdir -p ${base}/roles/$(role)/meta
	@echo "---" > ${base}/roles/$(role)/handlers/main.yaml
	@echo "---" > ${base}/roles/$(role)/templates/main.yaml
	@echo "---" > ${base}/roles/$(role)/files/main.yaml
	@echo "---" > ${base}/roles/$(role)/vars/main.yaml
	@echo "---" > ${base}/roles/$(role)/defaults/main.yaml
	@echo "---" > ${base}/roles/$(role)/meta/main.yaml
endif
	@echo "Role $(role) created successfully"

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
