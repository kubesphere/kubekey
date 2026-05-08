# Project

A project stores task templates to be executed, consisting of a series of YAML files. For easy understanding and quick adoption, KubeKey's task orchestration follows the conventions of [Ansible](https://github.com/ansible/ansible).

## Directory Structure

```text
project/
├── playbooks/          # Optional: store playbook files
│   ├── playbook1.yaml
│   └── playbook2.yaml
├── playbook1.yaml      # Or place playbooks directly in the project root
├── playbook2.yaml
└── roles/
    ├── roleName1/
    └── roleName2/
```

- **[playbooks](002-playbook.md)**: Execution entry point, storing playbooks. A playbook can define multiple tasks or roles, which run in the defined order during execution.
- **[roles](003-role.md)**: Collection of roles. A role is a group of [tasks](004-task.md).

## Storage Locations

Projects can be stored in **builtin**, **local**, or **Git**.

### Builtin

Builtin projects are located in the `builtin` directory and are integrated into the KubeKey command.

```shell
kk precheck
```

Executes `playbooks/precheck.yaml` in the `builtin` directory.

### Local

```shell
kk run demo.yaml
```

Executes `demo.yaml` in the current directory.

### Git

```shell
kk run playbooks/demo.yaml \
  --project-addr="$(GIT_URL)" \
  --project-branch="$(GIT_BRANCH)"
```

Executes `playbooks/demo.yaml` on the specified Git URL and branch.
