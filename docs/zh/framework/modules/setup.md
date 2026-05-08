# setup 模块

采集当前主机信息，即 [playbook](../002-playbook.md) 中 `gather_facts` 的底层实现。

## 参数

无额外参数。

## 示例

**1. 在 playbook 中启用 gather_facts**

```yaml
- name: playbook
  hosts: localhost
  gather_facts: true
```

**2. 在 task 中显式调用 setup**

```yaml
- name: setup
  setup: {}
```
