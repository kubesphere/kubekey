# 模板语法

任务与模板中的表达式遵循 [Go template](https://pkg.go.dev/text/template) 规范，并引用 [Sprig](https://github.com/Masterminds/sprig) 做函数扩展。

## 自定义函数

### toYaml

将变量转换为 YAML 字符串。可选参数为缩进空格数。

```yaml
{{ .yaml_variable | toYaml }}
```

### fromYaml

将 YAML 字符串解析为变量。

```yaml
{{ .yaml_string | fromYaml }}
```

### ipInCIDR

返回 CIDR 范围内的所有 IP 列表（数组）。

```yaml
{{ .cidr_variable | ipInCIDR }}
```

### ipFamily

返回 IP 或 IP_CIDR 的地址族。取值为 `Invalid`、`IPv4`、`IPv6`。

```yaml
{{ .ip | ipFamily }}
```

### pow

幂运算。

```yaml
# 2 的 3 次方，即 2 ** 3
{{ 2 | pow 3 }}
```

### subtractList

数组差集：返回在第一个参数中存在、但在第二个参数中不存在的元素组成的新列表。

```yaml
{{ .b | subtractList .a }}
```

### fileExist

判断路径对应的文件是否存在。

```yaml
{{ .file_path | fileExist }}
```
