# fetch Module

The fetch module allows users to pull files from a remote host to the local machine.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| src       | Path of the file on the remote host to fetch | string | Yes | - |
| dest      | Path on the local machine to save the fetched file | string | Yes | - |

## Usage Examples

1. Fetch a file
```yaml
- name: fetch file
  fetch:
    src: /tmp/src.yaml
    dest: /tmp/dest.yaml
```