# include_vars Module

The include_vars module allows users to apply variables to the specified hosts.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| include_vars | Path to the referenced file, must be YAML/YML format | string | Yes | - |

## Usage Examples

1. Set string variables
```yaml
- name: set other var file
  include_vars: "{{ .os.architecture }}/var.yaml"
```
