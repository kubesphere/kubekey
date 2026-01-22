# Task

Tasks are divided into **single-layer** and **multi-layer** types:

- **Single-layer task**: Contains module-related fields, does not contain `block`. A task can only use one module.
- **Multi-layer task**: Does not contain module fields, contains `block` (with optional `rescue`, `always`).

Tasks execute on each host of the play separately (unless `run_once: true`).

## File Definition

```yaml
- include_tasks: other/task.yaml
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: { a: b }

- name: Block Name
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: { a: b }
  block:
    - name: Task Name
      [module]
  rescue:
    - name: Task Name
      [module]
  always:
    - name: Task Name
      [module]

- name: Task Name
  tags: ["always"]
  when: true
  loop: [""]
  [module]
```

| Field | Description |
|-------|-------------|
| **include_tasks** | Reference other task files. |
| **name** | Task name, optional. |
| **tags** | Tags, optional. Only applies to this task, does not inherit play/role tags. |
| **when** | Execution condition, optional. Can be a string or array, using [template syntax](101-syntax.md), evaluated separately for each host. |
| **failed_when** | Failure condition, optional. Considered failed when met, supports [template syntax](101-syntax.md). |
| **run_once** | Whether to execute only once, optional, default `false`. Executes on the first host. |
| **ignore_errors** | Whether to ignore failures, optional, default `false`. |
| **vars** | Variables for this task, optional, YAML format. |
| **loop** | Execute module in a loop, passing current value as `item` each iteration. Can be a string or array, using [template syntax](101-syntax.md). |
| **retries** | Number of retries on failure, optional. |
| **register** | Write execution result to [variable](201-variable.md) for subsequent tasks. Contains sub-fields like `stderr`, `stdout`. |
| **register_type** | Parse format for `register`: `string` (default), `json`, `yaml`. |
| **block** | Task list. Required when no module is defined, executes in normal flow. |
| **rescue** | Task list. Executes when any sibling task in `block` fails. |
| **always** | Task list. Executes after `block` (and `rescue` if present) regardless of success or failure. |
| **module** | Specific operation, corresponding to [registered modules](README.md#modules). Required when not using `block`. |

## Registered Modules

| Module | Description |
|--------|-------------|
| [add_hostvars](modules/add_hostvars.md) | Inject variables into specified hosts |
| [assert](modules/assert.md) | Conditional assertion |
| [command](modules/command.md) | Execute commands |
| [copy](modules/copy.md) | Copy files/directories to target hosts |
| [debug](modules/debug.md) | Print variables |
| [fetch](modules/fetch.md) | Fetch files from remote hosts to local |
| [gen_cert](modules/gen_cert.md) | Validate or generate certificates |
| [image](modules/image.md) | Pull/push/copy images |
| [include_vars](modules/include_vars.md) | Load variables from YAML files |
| [prometheus](modules/prometheus.md) | Query Prometheus metrics |
| [result](modules/result.md) | Write to playbook status detail |
| [set_fact](modules/set_fact.md) | Set variables on the current host |
| [setup](modules/setup.md) | Gather host information |
| [template](modules/template.md) | Render templates and copy to target hosts |
