# setup Module

The setup module is the underlying implementation of gather_fact, allowing users to retrieve information about hosts.

## Parameters

null

## Usage Examples

1. Use gather_fact in a playbook
```yaml
- name: playbook 
  hosts: localhost
  gather_fact: true
```

2. Use setup in a task
```yaml
- name: setup
  setup: {}
```     
