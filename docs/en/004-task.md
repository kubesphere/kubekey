# Task
Tasks are divided into single-level tasks and multi-level tasks.  
Single-level tasks: Contain module-related fields and do not contain the `block` field. A task can contain only one module.  
Multi-level tasks: Do not contain module-related fields and must contain the `block` field.  
When a task runs, it is executed separately on each defined host.  

## File Definition
```yaml
- include_tasks: other/task.yaml
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: {a: b}
  
- name: Block Name
  tags: ["always"]
  when: true
  run_once: false
  ignore_errors: false
  vars: {a: b}
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
**include_tasks**: References other task template files in this task.  
**name**: Task name, optional.  
**tags**: Task tags, optional. They apply only to the task itself; roles and playbooks do not inherit them.  
**when**: Execution condition, can be a single value (string) or multiple values (array). Optional. By default, the role is executed. Values use [template syntax](101-syntax.md) and are evaluated separately for each host.  
**failed_when**: Failure condition. When a host meets this condition, the task is considered failed. Can be a single value (string) or multiple values (array). Optional. Values use [template syntax](101-syntax.md) and are evaluated separately for each host.  
**run_once**: Whether to execute only once. Optional. Defaults to false. If true, it runs on the first host.  
**ignore_errors**: Whether to ignore failures. Optional. Defaults to false.  
**vars**: Defines default parameters. Optional. YAML format.  
**loop**: Executes the module operation repeatedly. On each iteration, the value is passed to the module as `item: loop-value`. Can be a single value (string) or multiple values (array). Optional. Values use [template syntax](101-syntax.md) and are evaluated separately for each host.  
**retries**: Number of times to retry the task if it fails.  
**register**: A string value that registers the task result into a [variable](201-variable.md), which can be used in subsequent tasks. The register contains two subfields:  
- stderr: Failure output  
- stdout: Success output  
**register_type**: Format for registering `stderr` and `stdout` in the register.  
- string: Default, registers as a string.  
- json: Registers as JSON.  
- yaml: Registers as YAML.  
**block**: A collection of tasks. Optional (required if no module-related fields are defined). Always runs.  
**rescue**: A collection of tasks. Optional. Runs when the block fails (if any task in the block fails, the block fails).  
**always**: A collection of tasks. Optional. Always runs after block and rescue, regardless of success or failure.  
**module**: The actual operation to execute. Optional (required if no `block` field is defined). A map where the key is the module name and the value is the arguments. Available modules must be registered in advance in the project. Registered modules include:  
- [add_hostvars](modules/add_hostvars.md)  
- [assert](modules/assert.md)  
- [command](modules/command.md)  
- [copy](modules/copy.md)  
- [debug](modules/debug.md)  
- [fetch](modules/fetch.md)  
- [gen_cert](modules/gen_cert.md)  
- [image](modules/image.md)  
- [prometheus](modules/prometheus.md)  
- [result](modules/result.md)  
- [set_fact](modules/set_fact.md)  
- [setup](modules/setup.md)  
- [template](modules/template.md)  
- [include_vars](modules/include_vars.md)  
```