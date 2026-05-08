# result Module

Write variables to the playbook's `status.detail` for displaying execution results.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| any | Any key-value, will be merged into `status.detail` | string or map | No | - |

When calling `result` multiple times, all results are merged; duplicate keys use the last value.

## Examples

**1. Set string**

```yaml
- name: set string
  result:
    a: b
    c: d
```

`status` example:

```yaml
status:
  detail:
    a: b
    c: d
  phase: Succeeded
```

**2. Set nested map**

```yaml
- name: set map
  result:
    a:
      b: c
```

**3. Multiple sets (merged)**

```yaml
- name: set result1
  result:
    k1: v1

- name: set result2
  result:
    k2: v2

- name: set result3
  result:
    k2: v3
```

Final `detail` is `k1: v1`, `k2: v3` (later `k2` overrides earlier).
