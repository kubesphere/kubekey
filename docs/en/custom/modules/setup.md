# setup Module

Gather current host information, which is the underlying implementation of `gather_facts` in [playbooks](../002-playbook.md).

## Parameters

No additional parameters.

## Examples

**1. Enable gather_facts in playbook**

```yaml
- name: playbook
  hosts: localhost
  gather_facts: true
```

**2. Explicitly call setup in task**

```yaml
- name: setup
  setup: {}
```
