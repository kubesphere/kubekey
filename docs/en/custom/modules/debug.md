# debug Module

Print variables or strings for troubleshooting.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| msg | Content to output, supports [template syntax](../101-syntax.md) | string | Yes | - |

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

**2. Print map**

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

**3. Print array**

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
