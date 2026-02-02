# set_fact 模块

在当前执行主机上设置变量，供后续 task 使用。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| any | 要设置的变量，键值对形式 | 字符串或 map | 否 | - |

## 示例

**1. 设置字符串**

```yaml
- name: set string
  set_fact:
    a: b
    c: d
```

**2. 设置嵌套 map**

```yaml
- name: set map
  set_fact:
    a:
      b: c
```
