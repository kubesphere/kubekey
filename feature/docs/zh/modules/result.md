# result 模块

result模块允许用户将变量设置到playbook的status detail中显示。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| any | 需要设置的任意参数 | 字符串或map | 否 | - |

## 使用示例

1. 设置字符串参数
```yaml
- name: set string
  result:
    a: b
    c: d
```
playbook中status显示为：
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    a: b
    c: d
  phase: Succeeded
```

2. 设置map参数
```yaml
- name: set map
  result:
    a: 
      b: c
```
playbook中status显示为：
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    a:
      b: c
  phase: Succeeded
```

3. 设置多个result
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
所有的结果都会合并，如果有重复的key值，以最后设置的key为准。
playbook中status显示为：
```yaml
apiVersion: kubekey.kubesphere.io/v1
kind: Playbook
status:
  detail:
    k1: v1
    k2: v3
  phase: Succeeded
```