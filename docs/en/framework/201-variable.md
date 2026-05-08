# Variables

Variables are divided into **static variables** (defined before execution) and **dynamic variables** (generated during execution).
Priority: **Dynamic variables > Static variables**.

## Static Variables

Includes: inventory, global config, and parameters defined in templates.
Priority: **Global config > Inventory > Parameters defined in templates**.

### Inventory

YAML format, **does not** support template syntax. Passed via `-i` (`kk -i inventory.yaml ...`), effective on each host.

**Specification Example**:

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    hostname1:
      k1: v1
    hostname2:
      k2: v2
    hostname3: {}
  groups:
    groupname1:
      groups:
        - groupname2
      hosts:
        - hostname1
      vars:
        k1: v1
    groupname2: {}
  vars:
    k1: v1
```

| Field | Description |
|-------|-------------|
| **hosts** | Key is host name, value is variables for that host. |
| **groups** | Group hosts. Key is group name, value contains `groups`, `hosts`, `vars`. `groups` are nested groups, `hosts` are hosts included in the group, `vars` are group variables that apply to all hosts in the group. Host set in group = hosts in `groups` ∪ hosts in `hosts`. |
| **vars** | Global variables, apply to all hosts. |

Variable priority: **Host variables > Group variables > Global variables**.

### Global Config

YAML format, **does not** support template syntax. Passed via `-c` (`kk -c config.yaml ...`), effective on each host.

```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Config
metadata:
  name: default
spec:
  k: v
```

Any key-value can be defined under `spec`, participating in variable resolution along with [inventory](#inventory).

### Parameters Defined in Templates

Includes:

- `vars`, `vars_files` of playbook
- `defaults/main.yml`, `vars` of role
- `vars` of task

## Dynamic Variables

Generated at runtime, includes:

- **gather_facts** gathered host information
- **register** registered task output
- **set_fact** set variables

With the same name, **later definitions override earlier ones**.
