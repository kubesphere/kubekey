# result 模块

将变量写入 playbook 的 `status.detail`，用于展示执行结果。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| any | 任意键值，会合并到 `status.detail` | 字符串或 map | 否 | - |

多次调用 `result` 时，所有结果会合并；重复的 key 以最后一次为准。

## 示例

**1. 设置字符串**

```yaml
- name: set string
  result:
    a: b
    c: d
```

`status` 示例：

```yaml
status:
  detail:
    a: b
    c: d
  phase: Succeeded
```

**2. 设置嵌套 map**

```yaml
- name: set map
  result:
    a:
      b: c
```

**3. 多次设置（合并）**

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

最终 `detail` 为 `k1: v1`、`k2: v3`（后设置的 `k2` 覆盖前面的）。
