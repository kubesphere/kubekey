# add_hostvars Module

Inject variables into **specified hosts** for use by subsequent tasks.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| hosts | Target hosts, host name or group name | string or string array | No | - |
| vars | Variables to set | map | No | - |

## Examples

**1. Set string variable**

```yaml
- name: set string
  add_hostvars:
    hosts: all
    vars:
      c: d
```

**2. Set nested variable**

```yaml
- name: set map
  add_hostvars:
    hosts: all
    vars:
      a:
        b: c
```
