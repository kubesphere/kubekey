# include_vars 模块

include_vars模块允许用户将变量设置到指定的主机中生效。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| include_vars | 引用的文件地址，类型必须是yaml/yml | 字符串 | 是 | - |

## 使用示例

1. 设置字符串参数
```yaml
- name: set other var file
  include_vars: "{{ .os.architecture }}/var.yaml"
```

