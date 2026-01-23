# Template Syntax

Expressions in tasks and templates follow [Go template](https://pkg.go.dev/text/template) conventions, with [Sprig](https://github.com/Masterminds/sprig) function extensions.

## Custom Functions

### toYaml

Converts a variable to a YAML string. Optional parameter is the number of indentation spaces.

```yaml
{{ .yaml_variable | toYaml }}
```

### fromYaml

Parses a YAML string into a variable.

```yaml
{{ .yaml_string | fromYaml }}
```

### ipInCIDR

Returns all IP addresses (as an array) within the CIDR range.

```yaml
{{ .cidr_variable | ipInCIDR }}
```

### ipFamily

Returns the address family of an IP or IP_CIDR. Values: `Invalid`, `IPv4`, `IPv6`.

```yaml
{{ .ip | ipFamily }}
```

### pow

Power operation.

```yaml
# 2 to the power of 3, i.e., 2 ** 3
{{ 2 | pow 3 }}
```

### subtractList

Array difference: returns a new list containing elements that exist in the first argument but not in the second.

```yaml
{{ .b | subtractList .a }}
```

### fileExist

Checks if a file exists at the given path.

```yaml
{{ .file_path | fileExist }}
```
