# debug Module

The debug module lets users print variables.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| msg       | Content to print | string | Yes | - |

## Usage Examples

1. Print a string
```yaml
- name: debug string
  debug:
    msg: I'm {{ .name }}
```
If the variable `name` is `kubekey`, the output will be:
```txt
DEBUG: 
I'm kubekey
```

2. Print a map
```yaml
- name: debug map
  debug:
    msg: >-
      {{ .product }}
```
If the variable `product` is a map, e.g., `{"name":"kubekey"}`, the output will be:
```txt
DEBUG: 
{
    "name": "kubekey"
}
```

3. Print an array
```yaml
- name: debug array
  debug:
    msg: >-
      {{ .version }}
```
If the variable `version` is an array, e.g., `["1","2"]`, the output will be:
```txt
DEBUG: 
[
    "1",
    "2"
]
```