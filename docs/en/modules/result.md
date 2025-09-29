# result Module

The result module allows users to set variables to be displayed in the playbook's status detail.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| any       | Any parameter to set | string or map | No | - |

## Usage Examples

1. Set string parameters
```yaml
- name: set string
  result:
    a: b
    c: d
```
The status in the playbook will show:
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    a: b
    c: d
  phase: Succeeded
```

2. Set map parameters
```yaml
- name: set map
  result:
    a: 
      b: c
```
The status in the playbook will show:
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    a:
      b: c
  phase: Succeeded
```

3. Set multiple results
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
All results will be merged. If there are duplicate keys, the last set key will take precedence.
The status in the playbook will show:
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    k1: v1
    k2: v3
  phase: Succeeded
```