# assert Module

Assert conditions; task fails if conditions are not met.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| that | Assertion conditions, supports [template syntax](../101-syntax.md) | array or string | Yes | - |
| success_msg | Message to stdout when assertion is true | string | No | True |
| fail_msg | Message to stderr when assertion is false | string | No | False |
| msg | Same as fail_msg, lower priority than fail_msg | string | No | False |

## Examples

**1. Single condition (string)**

```yaml
- name: assert single condition
  assert:
    that: eq 1 1
```

- stdout: `"True"`
- stderr: `""`

**2. Multiple conditions (array)**

```yaml
- name: assert multi-condition
  assert:
    that:
      - eq 1 1
      - eq 1 2
```

- stdout: `"False"`
- stderr: `"False"`

**3. Custom success message**

```yaml
- name: assert is succeed
  assert:
    that: eq 1 1
    success_msg: "It's succeed"
```

- stdout: `"It's succeed"`
- stderr: `""`

**4. Custom failure message**

```yaml
- name: assert is failed
  assert:
    that: eq 1 2
    fail_msg: "It's failed"
    msg: "It's failed!"
```

- stdout: `"False"`
- stderr: `"It's failed"`
