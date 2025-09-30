# command (shell) Module

The command or shell module allows users to execute specific commands. The type of command executed is determined by the corresponding connector implementation.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| command   | The command to execute. Template syntax can be used. | string | yes | - |

## Usage Examples

1. Execute a shell command  
Connector type is `local` or `ssh`:
```yaml
- name: execute shell command
  command: echo "aaa"
```

2. Execute a Kubernetes command  
Connector type is `kubernetes`:
```yaml
- name: execute kubernetes command
  command: kubectl get pod
```
