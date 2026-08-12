# Inject Playbook (inject_playbooks.yaml)

![architecture](../../../images/architecture.png)

`inject_playbooks.yaml` is an **example playbook (template) for injection**. It is a no-op by
default, used to override/modify parameters after the default parameters are loaded, without
changing builtin playbooks or role code. KubeKey does not bundle this file; create it in your
own project (or local playbook directory).

After creating it, reference it via the
[`playbooks` injection mechanism](../../framework/002-playbook.md#inject-playbooks):

```yaml
spec:
  playbooks:
    - order: 1.5                        # insert between the 1st and 2nd original plays
      path: hook/inject_playbooks.yaml # the injection playbook created in your local project
```

## What you can do

- Use `set_fact` to override runtime variables of the current host (for a single host play).
- Use `add_hostvars` to write/override variables on multiple hosts in bulk (like `set_fact`
  but across hosts).
- Use `debug` to print variables and verify the injection took effect.

## Full example

The file content is commented out by default; uncomment the relevant block to enable it:

```yaml
- name: Inject | Inject and override playbook variables
  hosts:
    - all
  tasks:
    # Example 1: override a single variable (applies to all hosts of the current play)
    # - name: Inject | Override a single variable via set_fact
    #   set_fact:
    #     .kubernetes.version: "v1.31.0"

    # Example 2: override variables by group / host condition
    # - name: Inject | Override variables for a specific group
    #   set_fact:
    #     .kubernetes.container_manager: "containerd"
    #   when:
    #     - .groups.k8s_cluster | default list | has .inventory_hostname

    # Example 3: inject variables to multiple hosts in bulk (add_hostvars)
    # - name: Inject | Add host variables to multiple hosts
    #   add_hostvars:
    #     hosts: ["all"]
    #     vars:
    #       custom_label: "injected-by-hook"

    # Example 4: debug — confirm variable values before injection
    # - name: Inject | Debug current variables
    #   debug:
    #     msg: "kubernetes version is {{ .kubernetes.version }}"
```

## Notes

- Injected variables are only valid **during the current playbook run** (runtime variables)
  and are not written back to your config file. For persistent customization, prefer
  declaring it in the config spec.
- Only enable the example when you truly need "recompute / conditional override after the
  default parameters are loaded".
- Create this file in your own project (or local playbook directory) and reference it by a
  relative path; the KubeKey builtin package no longer ships it.
- For `order` positioning, sorting rules, empty-path skipping, etc., see
  [Inject Playbooks](../../framework/002-playbook.md#inject-playbooks).
