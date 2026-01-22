# fetch Module

Fetch files from remote hosts to local.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| src | File path on remote host | string | Yes | - |
| dest | Local save path | string | Yes | - |

## Examples

**1. Fetch file**

```yaml
- name: fetch file
  fetch:
    src: /tmp/src.yaml
    dest: /tmp/dest.yaml
```
