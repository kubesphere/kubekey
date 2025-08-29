# 语法
语法遵循`go template`规范.引用[sprig](https://github.com/Masterminds/sprig)进行函数扩展.

# 自定义函数

## toYaml
将参数转换成yaml字符串. 参数为左移空格数, 值为字符串
```yaml
{{ .yaml_variable | toYaml }}
```

## fromYaml
将yaml字符串转成参数格式
```yaml
{{ .yaml_string | fromYaml }}
```

## ipInCIDR
获取IP范围(cidr)内的所有ip列表(数组)
```yaml
{{ .cidr_variable | ipInCIDR }}
```

## ipFamily
获取IP或IP_CIDR所属的family。返回值为Invalid, IPv4, IPv6 
```yaml
{{ .ip | ipFamily }}
```

## pow
幂运算.
```yaml
# 2的3次方, 2 ** 3
{{ 2 | pow 3 }}
```

## subtractList
数组不包含
```yaml
# 返回一个新列表，该列表中的元素在a中存在，但在b中不存在
{{ .b | subtractList .a }}
```

## fileExist
数组不包含
```yaml
# 判断文件是否存在
{{ .file_path | fileExist }}
```