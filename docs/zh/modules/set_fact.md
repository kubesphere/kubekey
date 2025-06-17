# set_fact 模块

set_fact模块允许用户将变量设置到所有的主机中生效。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| any | 需要设置的任意参数 | 字符串或map | 否 | - |

## 使用示例

1. 设置字符串参数
```yaml
- name: set string
  set_fact:
    a: b
    c: d
```

2. 设置map参数
```yaml
- name: set map
  set_fact:
    a: 
      b: c
```
