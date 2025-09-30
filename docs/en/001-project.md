# Project
The project stores task templates to be executed, consisting of a series of YAML files.  
To help users quickly understand and get started, kk’s task abstraction is inspired by [ansible](https://github.com/ansible/ansible)’s playbook specification.

## Directory Structure
```text
|-- project
|   |-- playbooks/  
|   |-- playbook1.yaml  
|   |-- playbook2.yaml  
|   |-- roles/
|   |   |-- roleName1/    
|   |   |-- roleName2/    
...
```
**[playbooks](002-playbook.md)**: The execution entry point. Stores a series of playbooks. A playbook can define multiple tasks or roles. When running a workflow template, the defined tasks are executed in order.  
**[roles](003-role.md)**: A collection of roles. A role is a group of tasks.

## Storage Locations
Projects can be stored as built-in, local, or on a Git server.  

### Built-in
Built-in projects are stored in the `builtin` directory and integrated into kubekey commands.  
Example:
```shell
kk precheck
```
This runs the `playbooks/precheck.yaml` workflow file in the `builtin` directory.  

### Local
Example:
```shell
kk run demo.yaml
```
This runs the `demo.yaml` workflow file in the current directory.  

### Git
Example:
```shell
kk run playbooks/demo.yaml \
  --project-addr=$(GIT_URL) \
  --project-branch=$(GIT_BRANCH)
```
This runs the `playbooks/demo.yaml` workflow file from the Git repository at `$(GIT_URL)`, branch `$(GIT_BRANCH)`.  
