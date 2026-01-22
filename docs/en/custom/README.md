# Custom Project Development

This document explains how to write and run custom playbook projects based on KubeKey. KubeKey's task orchestration follows the conventions of [Ansible](https://github.com/ansible/ansible) for easy understanding and quick adoption.

## Documentation Navigation

### Basic Concepts

| Document | Description |
|----------|-------------|
| [Project (001-project)](001-project.md) | Project structure, playbooks/roles directories, builtin/local/Git storage methods |
| [Playbook (002-playbook)](002-playbook.md) | Playbook definition, hosts/tags/serial, pre_tasks/roles/tasks/post_tasks |
| [Role (003-role)](003-role.md) | Role structure, defaults/tasks/templates/files, referencing in playbooks |
| [Task (004-task)](004-task.md) | Task definition, single/multi-layer task, block/rescue/always, loop/register |

### Syntax and Variables

| Document | Description |
|----------|-------------|
| [Template Syntax (101-syntax)](101-syntax.md) | Go template and Sprig, toYaml/fromYaml, ipInCIDR, custom functions |
| [Variables (201-variable)](201-variable.md) | Static variables (inventory, global config, template params) and dynamic variables (gather_facts, register, set_fact) |

### Modules

The modules available in tasks must be registered in the project. The following modules are already registered:

| Module | Description |
|--------|-------------|
| [add_hostvars](modules/add_hostvars.md) | Inject variables into specified hosts |
| [assert](modules/assert.md) | Conditional assertion |
| [command](modules/command.md) | Execute commands (shell/kubectl, etc.) |
| [copy](modules/copy.md) | Copy files or directories to target hosts |
| [debug](modules/debug.md) | Print variables |
| [fetch](modules/fetch.md) | Fetch files from remote hosts to local |
| [gen_cert](modules/gen_cert.md) | Validate or generate certificates |
| [image](modules/image.md) | Pull/push/copy images |
| [include_vars](modules/include_vars.md) | Load variables from YAML files |
| [prometheus](modules/prometheus.md) | Query Prometheus metrics |
| [result](modules/result.md) | Write to playbook status detail |
| [set_fact](modules/set_fact.md) | Set variables on the current host |
| [setup](modules/setup.md) | Gather host information (gather_facts underlying) |
| [template](modules/template.md) | Render templates and copy to target hosts |

## Quick Start

1. Read [Project](001-project.md) to understand the directory structure and storage methods.
2. Read [Playbook](002-playbook.md) and [Task](004-task.md) to write your first playbook and task.
3. Use [Template Syntax](101-syntax.md) and [Variables](201-variable.md) to reference and pass parameters.
4. Refer to the [Module](modules/) documentation as needed and select the appropriate module for your specific logic.
