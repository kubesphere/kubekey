# Syntax
The syntax follows the `go template` specification, with function extensions provided by [sprig](https://github.com/Masterminds/sprig).

# Custom Functions

## toYaml
Converts a parameter into a YAML string. The argument specifies the number of leading spaces, and the value is a string.
```yaml
{{ .yaml_variable | toYaml }}
```

## fromYaml
Converts a YAML string into a parameter format.
```yaml
{{ .yaml_string | fromYaml }}
```

## ipInCIDR
Gets all IP addresses (as an array) within the specified IP range (CIDR).
```yaml
{{ .cidr_variable | ipInCIDR }}
```

## ipFamily
Determines the family of an IP or IP_CIDR. Returns: Invalid, IPv4, or IPv6.
```yaml
{{ .ip | ipFamily }}
```

## pow
Performs exponentiation.
```yaml
# 2 to the power of 3, 2 ** 3
{{ 2 | pow 3 }}
```

## subtractList
Array exclusion.
```yaml
# Returns a new list containing elements that exist in a but not in b
{{ .b | subtractList .a }}
```

## fileExist
Checks if a file exists.
```yaml
{{ .file_path | fileExist }}
```