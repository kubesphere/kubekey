# Role

A Role is a collection of [tasks](004-task.md).

## Referencing in Playbooks

```yaml
- name: Playbook Name
  # ...
  roles:
    - name: Role Name
      tags: ["always"]
      when: true
      run_once: false
      ignore_errors: false
      vars: { a: b }
      role: Role-ref Name
```

| Field | Description |
|-------|-------------|
| **name** | Role display name, optional. Can be different from the `role` reference name. |
| **tags** | Tags, optional. Only applies to this role reference. |
| **when** | Execution condition, optional. Can be a string or array, evaluated separately for each host. |
| **run_once** | Whether to execute only once, optional, default `false`. Executes on the first host. |
| **ignore_errors** | Whether to ignore task failures under this role, optional, default `false`. |
| **role** | Reference name, required. Corresponds to the subdirectory name under `roles/`. |
| **vars** | Default variables, optional, YAML format. |

## Role Directory Structure

```text
project/roles/roleName/
├── defaults/
│   └── main.yml    # Default variables, applies to all tasks under this role
├── tasks/
│   └── main.yml    # [Task](004-task.md) definition
├── templates/      # Template files, for template-type tasks
│   └── template1
└── files/          # Static files, for copy-type tasks
    └── file1
```

- **roleName**: The reference name in playbooks' `role`, can be multi-level directories (e.g., `a/b`).
- **defaults**: Define default parameters for this role in `main.yml`.
- **tasks**: Define the [task](004-task.md) list for this role in `main.yml`.
- **templates**: Template files, usually containing [template syntax](101-syntax.md) variable references.
- **files**: Raw files, referenced by relative path in [copy](modules/copy.md) tasks.
