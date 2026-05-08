# assert 模块

对条件做断言，不满足时任务失败。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| that | 断言条件，支持 [模板语法](../101-syntax.md) | 数组或字符串 | 是 | - |
| success_msg | 断言为 true 时写入 stdout 的信息 | 字符串 | 否 | True |
| fail_msg | 断言为 false 时写入 stderr 的信息 | 字符串 | 否 | False |
| msg | 同 fail_msg，优先级低于 fail_msg | 字符串 | 否 | False |

## 示例

**1. 单个条件（字符串）**

```yaml
- name: assert single condition
  assert:
    that: eq 1 1
```

- stdout: `"True"`
- stderr: `""`

**2. 多个条件（数组）**

```yaml
- name: assert multi-condition
  assert:
    that:
      - eq 1 1
      - eq 1 2
```

- stdout: `"False"`
- stderr: `"False"`

**3. 自定义成功信息**

```yaml
- name: assert is succeed
  assert:
    that: eq 1 1
    success_msg: "It's succeed"
```

- stdout: `"It's succeed"`
- stderr: `""`

**4. 自定义失败信息**

```yaml
- name: assert is failed
  assert:
    that: eq 1 2
    fail_msg: "It's failed"
    msg: "It's failed!"
```

- stdout: `"False"`
- stderr: `"It's failed"`
