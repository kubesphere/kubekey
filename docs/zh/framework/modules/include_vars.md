# include_vars 模块

从 YAML 文件加载变量到当前执行上下文，供后续 task 使用。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| include_vars | 变量文件路径，须为 `.yaml` 或 `.yml` | 字符串 | 是 | - |

路径可使用 [模板语法](../101-syntax.md)，例如按架构加载不同文件。

## 示例

**1. 按条件加载变量文件**

```yaml
- name: load vars by architecture
  include_vars: "{{ .os.architecture }}/var.yaml"
```
