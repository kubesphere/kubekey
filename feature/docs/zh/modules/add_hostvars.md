# add_hostvars 模块

add_hostvars模块允许用户将变量设置到指定的主机中生效。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| hosts | 需要设置参数的目标主机 | 字符串或字符串数组 | 否 | - |
| vars | 需要设置的参数 | map | 否 | - |

## 使用示例

1. 设置字符串参数
```yaml
- name: set string
  add_hostvars:
    name: all
    vars:
      c: d
```

2. 设置map参数
```yaml
- name: set map
  add_hostvars:
    name: all
    vars:
      a: 
        b: c
```
