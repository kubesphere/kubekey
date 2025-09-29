# Playbook
## File Definition
A playbook file executes multiple playbooks in the defined order. Each playbook specifies which tasks to run on which hosts. 
```yaml
- import_playbook: others/playbook.yaml

- name: Playbook Name
  tags: ["always"]
  hosts: ["host1", "host2"]
  serial: 1
  run_once: false
  ignore_errors: false
  gather_facts: false
  vars: {a: b}
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
**import_playbooks**: Defines the referenced playbook file name, usually a relative path. File lookup order is: `project_path/playbooks/`, `current_path/playbooks/`, `current_path/`.  
**name**: Playbook name, optional.  
**tags**: Tags of the playbook, optional. They apply only to the playbook itself; roles and tasks under the playbook do not inherit these tags.  
When running a playbook command, you can filter which playbooks to execute using tags. For example:  
- `kk run [playbook] --tags tag1 --tags tag2`: Executes playbooks with either the tag1 or tag2 label.  
- `kk run [playbook] --skip-tags tag1 --skip-tags tag2`: Skips playbooks with the tag1 or tag2 label.  
Playbooks with the `always` tag always run. Playbooks with the `never` tag never run.  
When the argument is `all`, it selects all playbooks. When the argument is `tagged`, it selects only tagged playbooks.  
**hosts**: Defines which machines to run on. Required. All hosts must be defined in the `inventory` (except localhost). You can specify host names or group names.  
**serial**: Executes playbooks in batches. Can be a single value (string or number) or an array. Optional. Defaults to executing all at once.  
- If `serial` is an array of numbers, hosts are grouped by fixed sizes. If the hosts exceed the defined `serial` values, the last `serial` value is used.  
  For example, if `serial = [1, 2]` and `hosts = [a, b, c, d]`, the playbook runs in 3 batches: [a], [b, c], [d].  
- If `serial` is percentages, the number of hosts per batch is calculated based on percentages (rounded down). If hosts exceed the defined percentages, the last percentage is reused.  
  For example, if `serial = [30%, 60%]` and `hosts = [a, b, c, d]`, percentages translate to [1.2, 2.4], rounded to [1, 2].  
Numbers and percentages can be mixed.  
**run_once**: Whether to execute only once. Optional. Defaults to false. If true, it runs on the first host.  
**ignore_errors**: Whether to ignore failed tasks under this playbook. Optional. Defaults to false.  
**gather_facts**: Whether to gather system information. Optional. Defaults to false. Collects data per host.  
- localConnector: Collects release (/etc/os-release), kernel_version (uname -r), hostname (hostname), architecture (arch). Currently supports only Linux.  
- sshConnector: Collects release (/etc/os-release), kernel_version (uname -r), hostname (hostname), architecture (arch). Currently supports only Linux.  
- kubernetesConnector: Not supported yet.  
**vars**: Defines default parameters. Optional. YAML format.  
**vars_files**: Defines default parameters. Optional. YAML file format. Fields in `vars` and `vars_files` must not overlap.  
**pre_tasks**: Defines [tasks](004-task.md) to run before roles. Optional.  
**roles**: Defines [roles](003-role.md) to run. Optional.  
**tasks**: Defines [tasks](004-task.md) to run. Optional.  
**post_tasks**: Defines [tasks](004-task.md) to run after roles and tasks. Optional.  

## Playbook Execution Order
- Across playbooks: Executed in the defined order. If an `import_playbook` is included, the referenced file is converted into playbooks.  
- Within the same playbook: Execution order is pre_tasks -> roles -> tasks -> post_tasks.  
If any task fails (excluding ignored failures), the playbook execution fails.  
