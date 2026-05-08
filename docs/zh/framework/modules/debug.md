# debug 模块

打印变量或字符串，便于排查问题。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| msg | 要输出的内容，支持带过滤器的 [模板语法](../101-syntax.md) | 字符串 | 否 | - |
| var | 要输出的变量路径，支持模板语法 `{{ .var }}` 或简单路径 `.var` | 字符串 | 否 | - |

**注意：** `msg` 和 `var` 至少需要一个。

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

**2. 使用模板过滤器打印**

```yaml
- name: debug with default filter
  debug:
    msg: "Name is {{ .name | default \"unknown\" }}"
```

当 `name` 未定义时，输出：

```text
DEBUG:
Name is unknown
```

**3. 打印 map**

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

**4. 打印数组**

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

**5. 使用 var 字段打印变量（模板语法）**

```yaml
- name: debug variable with template
  debug:
    var: "{{ .config.version }}"
```

当 `config.version` 为 `v1.0.0` 时，输出：

```text
DEBUG:
"v1.0.0"
```

**6. 使用 var 字段打印变量（简单路径）**

```yaml
- name: debug variable with path
  debug:
    var: ".config.version"
```

当 `config.version` 为 `v1.0.0` 时，输出：

```text
DEBUG:
"v1.0.0"
```

**7. 打印嵌套变量**

```yaml
- name: debug nested variable
  debug:
    var: ".data.level1.level2"
```

当 `data.level1.level2` 为 `value` 时，输出：

```text
DEBUG:
"value"
```
