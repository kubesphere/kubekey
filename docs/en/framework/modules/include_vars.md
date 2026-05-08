# include_vars Module

Load variables from YAML files into the current execution context for use by subsequent tasks.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| include_vars | Variable file path, must be `.yaml` or `.yml` | string | Yes | - |

Paths can use [template syntax](../101-syntax.md), for example, loading different files based on architecture.

## Examples

**1. Load variable files conditionally**

```yaml
- name: load vars by architecture
  include_vars: "{{ .os.architecture }}/var.yaml"
```
