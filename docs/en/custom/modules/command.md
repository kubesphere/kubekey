# command / shell Module

Execute commands; specific behavior is determined by the connector type (e.g., `local`/`ssh` execute Shell, `kubernetes` executes kubectl, etc.).

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| command | Command to execute. Can use [template syntax](../101-syntax.md) | string | Yes | - |

## Usage Examples

**1. Execute Shell command** (connector is `local` or `ssh`)

```yaml
- name: execute shell command
  command: echo "aaa"
```

**2. Execute Kubernetes command** (connector is `kubernetes`)

```yaml
- name: executor kubernetes command
  command: kubectl get pod
```
