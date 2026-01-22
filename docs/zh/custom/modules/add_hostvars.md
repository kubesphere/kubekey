# add_hostvars 模块

向**指定主机**注入变量，供后续 task 使用。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| hosts | 目标主机，host 名或 group 名 | 字符串或字符串数组 | 否 | - |
| vars | 要设置的变量 | map | 否 | - |

## 示例

**1. 设置字符串变量**

```yaml
- name: set string
  add_hostvars:
    hosts: all
    vars:
      c: d
```

**2. 设置嵌套变量**

```yaml
- name: set map
  add_hostvars:
    hosts: all
    vars:
      a:
        b: c
```
