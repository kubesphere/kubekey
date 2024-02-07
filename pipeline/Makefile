.PHONY: create-role
create-role: ## create a role necessary file in roles
	@echo "Creating role $(role)..."
	@mkdir -p roles/$(role)/tasks
	@echo "---" > roles/$(role)/tasks/main.yaml
ifeq ($(VARIABLE_NAME),"full")
	@mkdir -p roles/$(role)/handlers
	@mkdir -p roles/$(role)/templates
	@mkdir -p roles/$(role)/files
	@mkdir -p roles/$(role)/vars
	@mkdir -p roles/$(role)/defaults
	@mkdir -p roles/$(role)/meta
	@echo "---" > roles/$(role)/handlers/main.yaml
	@echo "---" > roles/$(role)/templates/main.yaml
	@echo "---" > roles/$(role)/files/main.yaml
	@echo "---" > roles/$(role)/vars/main.yaml
	@echo "---" > roles/$(role)/defaults/main.yaml
	@echo "---" > roles/$(role)/meta/main.yaml
endif
	@echo "Role $(role) created successfully"

