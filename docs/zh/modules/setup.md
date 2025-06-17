# setup 模块

gather_fact的底层实现，setup模块允许用户获取主机的信息

## 参数

null

## 使用示例

1. 在playbook中使用gather_fact
```yaml
- name: playbook 
  hosts: localhost
  gather_fact: true
```

2. 在task中使用setup
```yaml
- name: setup
  setup: {}
```     
