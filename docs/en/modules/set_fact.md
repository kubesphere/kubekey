# set_fact Module

The set_fact module allows users to set variables effective on the currently executing host.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| any       | Any parameter to set | string or map | No | - |

## Usage Examples

1. Set string variables
```yaml
- name: set string
  set_fact:
    a: b
    c: d
```

2. Set map variables
```yaml
- name: set map
  set_fact:
    a: 
      b: c
```
