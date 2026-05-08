# command / shell 模块

执行命令，具体行为由 connector 类型决定（如 `local` / `ssh` 执行 Shell，`kubernetes` 执行 kubectl 等）。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| command | 执行的命令.可使用[模板语法](../101-syntax.md) | 字符串 | 是 | - |

## 使用示例

**1. 执行 Shell 命令**（connector 为 `local` 或 `ssh`）

```yaml
- name: execute shell command
  command: echo "aaa"
```

**2. 执行 Kubernetes 命令**（connector 为 `kubernetes`）

```yaml
- name: executor kubernetes command
  command: kubectl get pod
```
