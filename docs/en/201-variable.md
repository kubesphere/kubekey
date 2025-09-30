# Variables
Variables are divided into static variables (defined before execution) and dynamic variables (generated at runtime).  
The parameter priority is: dynamic variables > static variables.

## Static Variables
Static variables include the inventory, global configuration, and parameters defined in templates.  
The parameter priority is: global configuration > inventory > parameters defined in templates.

### Inventory
A YAML file without template syntax, passed in via the `-i` parameter (`kk -i inventory.yaml ...`), effective on each host.  
**Definition format**:
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    hostname1: 
      k1: v1
      #...
    hostname2: 
      k2: v2
      #...
    hostname3:
      #...
  groups:
    groupname1:
      groups:
        - groupname2
        # ...
      hosts:
        - hostname1
        #...
      vars:
        k1: v1
        #...
    groupname2:
    #...
  vars:
    k1: v1
    #...
```
**hosts**: The key is the host name, the value is the variables assigned to that host.  
**groups**: Defines host groups. The key is the group name, and the value includes groups, hosts, and vars.  
- groups: Other groups included in this group.  
- hosts: Hosts included in this group.  
- vars: Group-level variables, effective for all hosts in the group.  
The total hosts in a group are the sum of those in `groups` and those listed under `hosts`.  
**vars**: Global variables, effective for all hosts.  
Variable priority: $(host_variable) > $(group_variable) > $(global_variable).

### Global Configuration
A YAML file without template syntax, passed in via the `-c` parameter (`kk -c config.yaml ...`), effective on each host.  
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
metadata:
  name: default
spec:
  k: v
  #...
```
Parameters can be of any type.

### Parameters Defined in Templates
Parameters defined in templates include:  
- Parameters defined in the `vars` and `vars_files` fields in playbooks.  
- Parameters defined in `defaults/main.yaml` in roles.  
- Parameters defined in the `vars` field in roles.  
- Parameters defined in the `vars` field in tasks.  

## Dynamic Variables
Dynamic variables are generated during node execution and include:  
- Parameters defined by `gather_facts`.  
- Parameters defined by `register`.  
- Parameters defined by `set_fact`.  
Priority follows the order of definition: later definitions override earlier ones.  