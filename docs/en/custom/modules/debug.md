# debug Module

Print variables or strings for troubleshooting.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| msg | Content to output, supports [template syntax](../101-syntax.md) with filters | string | No | - |
| var | Variable path to output, supports template syntax `{{ .var }}` or simple path `.var` | string | No | - |

**Note:** One of `msg` or `var` must be specified.

## Examples

**1. Print string**

```yaml
- name: debug string
  debug:
    msg: I'm {{ .name }}
```

When `name` is `kubekey`, output:

```text
DEBUG:
I'm kubekey
```

**2. Print with template filter**

```yaml
- name: debug with default filter
  debug:
    msg: "Name is {{ .name | default \"unknown\" }}"
```

When `name` is not defined, output:

```text
DEBUG:
Name is unknown
```

**3. Print map**

```yaml
- name: debug map
  debug:
    msg: >-
      {{ .product }}
```

When `product` is `{"name":"kubekey"}`, output similar to:

```text
DEBUG:
{
    "name": "kubekey"
}
```

**4. Print array**

```yaml
- name: debug array
  debug:
    msg: >-
      {{ .version }}
```

When `version` is `["1","2"]`, output similar to:

```text
DEBUG:
[
    "1",
    "2"
]
```

**5. Print variable using var field (template syntax)**

```yaml
- name: debug variable with template
  debug:
    var: "{{ .config.version }}"
```

When `config.version` is `v1.0.0`, output:

```text
DEBUG:
"v1.0.0"
```

**6. Print variable using var field (simple path)**

```yaml
- name: debug variable with path
  debug:
    var: ".config.version"
```

When `config.version` is `v1.0.0`, output:

```text
DEBUG:
"v1.0.0"
```

**7. Print nested variable**

```yaml
- name: debug nested variable
  debug:
    var: ".data.level1.level2"
```

When `data.level1.level2` is `value`, output:

```text
DEBUG:
"value"
```
