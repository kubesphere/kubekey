# Playbook

## File Definition

A playbook file can execute multiple plays in the defined order; each play specifies which hosts to execute which tasks on.

```yaml
- import_playbook: others/playbook.yaml

- name: Playbook Name
  tags: ["always"]
  hosts: ["host1", "host2"]
  serial: 1
  run_once: false
  ignore_errors: false
  gather_facts: false
  vars: { a: b }
  vars_files: ["vars/variables.yaml"]
  pre_tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
  roles:
    - role: role1
      when: true
  tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
  post_tasks:
    - name: Task Name
      debug:
        msg: "I'm Task"
```

| Field | Description |
|-------|-------------|
| **import_playbook** | Path to the referenced playbook (usually relative). Search order: `project path/playbooks/` → `current path/playbooks/` → `current path/`. |
| **name** | Play name, optional. |
| **tags** | Tags for the play, optional. Only applies to that play and does not inherit to roles/tasks below. Can be filtered with `--tags` / `--skip-tags` during execution. `always` always executes, `never` never executes, `all` means all plays, `tagged` means tagged plays. |
| **hosts** | Execution target, required. Can be host names or group names, all must be defined in the [inventory](201-variable.md#inventory) (except localhost). |
| **serial** | Batch execution. Can be a single value (number or string) or an array. Default is one batch. If an array, `hosts` are grouped by fixed quantity; exceeding values extend with the last value. E.g., `[1, 2]`, `hosts: [a,b,c,d]` → first batch `[a]`, second batch `[b,c]`, third batch `[d]`. Supports percentages (e.g., `[30%, 60%]`), can be mixed with numbers. |
| **run_once** | Whether to execute only once, optional, default `false`. When `true`, executes on the first host. |
| **ignore_errors** | Whether to ignore task failures under this play, optional, default `false`. |
| **gather_facts** | Whether to gather host information, optional, default `false`. Gathers different data based on connector type (e.g., `local`/`ssh`: `release`, `kernel_version`, `hostname`, `architecture`, Linux only). |
| **vars** | Default variables, optional, YAML format. |
| **vars_files** | Load default variables from YAML files, optional. Keys cannot duplicate with `vars`. |
| **pre_tasks** | Pre-[tasks](004-task.md), optional. |
| **roles** | [Roles](003-role.md) to execute, optional. |
| **tasks** | Main [tasks](004-task.md), optional. |
| **post_tasks** | Post-[tasks](004-task.md), optional. |

## Execution Order

- **Multiple plays**: Execute in defined order; `import_playbook` expands to the corresponding play first.
- **Within the same play**: `pre_tasks` → `roles` → `tasks` → `post_tasks`.
- Any task failure (without `ignore_errors`) results in play failure.
