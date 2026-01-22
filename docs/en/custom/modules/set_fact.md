# set_fact Module

Set variables on the current execution host for use by subsequent tasks.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| any | Variables to set, in key-value format | string or map | No | - |

## Examples

**1. Set string**

```yaml
- name: set string
  set_fact:
    a: b
    c: d
```

**2. Set nested map**

```yaml
- name: set map
  set_fact:
    a:
      b: c
```
