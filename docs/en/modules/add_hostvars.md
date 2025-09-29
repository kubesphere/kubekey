# add_hostvars Module

The add_hostvars module allows users to set variables that take effect on the specified hosts.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| hosts     | Target hosts to set parameters for | String or array of strings | No | - |
| vars      | Parameters to set | Map | No | - |

## Usage Examples

1. Set a string parameter
```yaml
- name: set string
  add_hostvars:
    hosts: all
    vars:
      c: d
```

2. Set a map parameter
```yaml
- name: set map
  add_hostvars:
    hosts: all
    vars:
      a: 
        b: c
```