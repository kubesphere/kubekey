.PHONY: create-role
create-role: ## create a role necessary file in roles
	@echo "Creating role $(role)..."
	@mkdir -p roles/$(role)/tasks
	@echo "---" > roles/$(role)/tasks/main.yaml
	@mkdir -p roles/$(role)/defaults
	@echo "" > roles/$(role)/defaults/main.yaml
ifeq ($(VARIABLE_NAME),"full")
	@mkdir -p roles/$(role)/handlers
	@mkdir -p roles/$(role)/templates
	@mkdir -p roles/$(role)/files
	@mkdir -p roles/$(role)/vars
	@mkdir -p roles/$(role)/meta
	@echo "---" > roles/$(role)/handlers/main.yaml
	@echo "---" > roles/$(role)/templates/main.yaml
	@echo "---" > roles/$(role)/files/main.yaml
	@echo "---" > roles/$(role)/vars/main.yaml
	@echo "---" > roles/$(role)/defaults/main.yaml
	@echo "---" > roles/$(role)/meta/main.yaml
endif
	@echo "Role $(role) created successfully"

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
