# command(shell) 模块

command或shell模块允许用户执行特定命令。执行何种命令由相关的connector实现

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| command | 执行的命令.可使用[模板语法](../101-syntax.md) | 字符串 | 是 | - |

## 使用示例

1. 执行shell命令
connector.type 为 `local` 或 `ssh`
```yaml
- name: execute shell command
  command: echo "aaa"
```

1. 执行kubernetes命令
connector.type 为 `kubernetes`
```yaml
- name: executor kubernetes command
  command: kubectl get pod
```
