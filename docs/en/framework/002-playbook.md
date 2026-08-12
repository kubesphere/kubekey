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

## Inject Playbooks

Besides hardcoding `import_playbook` inside a playbook file, you can declare a `playbooks`
list in the playbook's config spec to inject a **custom playbook at any position of the
top-level (first) source playbook**. This lets you override/modify parameters after the
default parameters are loaded, without touching builtin playbooks or role code.

### Configuration

```yaml
spec:
  playbooks:
    - order: 1.5                        # insert between the 1st and 2nd original plays
      path: hook/inject_playbooks.yaml # playbook to inject (relative to project root, templated)
    - order: 0                          # 0 means insert at the very front (before the 1st play)
      path: hook/pre_custom.yaml
```

### Rules

- **Original plays get weights `1, 2, 3, ...` automatically** (by document order, 1-based).
- **Injected items use an explicit `order`** (float, can be fractional or negative) to
  position the insertion; e.g. `order: 1.5` inserts between the 1st and 2nd original plays,
  `order: 0` inserts at the very front.
- **Sort comparator (deterministic; duplicate `order` does not error)**:
  1. `order` ascending;
  2. `playbooks` config items take precedence over file plays;
  3. within the same source, by definition order (config list order / file order).
- **`path` is rendered through templates**:
  - rendered to empty / unset variable → that item is **skipped** (no error);
  - path set but file not found → **errors** (as usual), to catch typos.
- This mechanism applies **only to the top-level (first) source playbook.yaml**; imported
  sub-playbooks are not injected again.

### Example

A typical use case is overriding variables after the default parameters are loaded.
Create your own injection playbook in your project (e.g. `hook/inject_playbooks.yaml`,
a no-op by default), then reference it via `path`. For the full copyable example and
field reference, see
[Inject Playbook (inject_playbooks.yaml)](../reference/playbooks/hook/inject_playbooks.md).
