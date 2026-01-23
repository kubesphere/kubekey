# debug 模块

打印变量或字符串，便于排查问题。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| msg | 要输出的内容，支持 [模板语法](../101-syntax.md) | 字符串 | 是 | - |

## 示例

**1. 打印字符串**

```yaml
- name: debug string
  debug:
    msg: I'm {{ .name }}
```

当 `name` 为 `kubekey` 时，输出：

```text
DEBUG:
I'm kubekey
```

**2. 打印 map**

```yaml
- name: debug map
  debug:
    msg: >-
      {{ .product }}
```

当 `product` 为 `{"name":"kubekey"}` 时，输出类似：

```text
DEBUG:
{
    "name": "kubekey"
}
```

**3. 打印数组**

```yaml
- name: debug array
  debug:
    msg: >-
      {{ .version }}
```

当 `version` 为 `["1","2"]` 时，输出类似：

```text
DEBUG:
[
    "1",
    "2"
]
```
