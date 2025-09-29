# Role
A role is a group of tasks.

## Defining a Role Reference in a Playbook
```yaml
- name: Playbook Name
  #...
  roles:
    - name: Role Name
      tags: ["always"]
      when: true
      run_once: false
      ignore_errors: false
      vars: {a: b}
      role: Role-ref Name
```
**name**: Role name, optional. This name is different from the role reference name in the playbook.  
**tags**: Tags of the playbook, optional. They apply only to the playbook itself; roles and tasks under it do not inherit these tags.  
**when**: Execution condition, can be a single value (string) or multiple values (array). Optional. By default, the role is executed. The condition is evaluated separately for each host.  
**run_once**: Whether to execute only once. Optional. Defaults to false. If true, it runs on the first host.  
**ignore_errors**: Whether to ignore failures of tasks under this role. Optional. Defaults to false.  
**role**: The reference name used in the playbook, corresponding to a subdirectory under the `roles` directory. Required.  
**vars**: Defines default parameters. Optional. YAML format.  

## Role Directory Structure
```text
|-- project
|   |-- roles/
|   |   |-- roleName/  
|   |   |   |-- defaults/  
|   |   |   |   |-- main.yml  
|   |   |   |-- tasks/  
|   |   |   |   |-- main.yml  
|   |   |   |-- templates/  
|   |   |   |   |-- template1  
|   |   |   |-- files/  
|   |   |   |   |-- file1    
```
**roleName**: The reference name of the role. Can be a single-level or multi-level directory.  
**defaults**: Defines default parameter values for all tasks under the role. Defined in the `main.yml` file.  
**[tasks](004-task.md)**: Task templates associated with the role. A role can include multiple tasks, defined in the `main.yml` file.  
**templates**: Template files, which usually reference variables. Used in tasks of type `templates`.  
**files**: Raw files, used in tasks of type `copy`.  
