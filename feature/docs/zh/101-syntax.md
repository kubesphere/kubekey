# 语法
语法遵循`go template`规范.引用[sprig](https://github.com/Masterminds/sprig)进行函数扩展.
# 自定义函数
## pow
幂运算.
```yaml
# 2的3次方, 2 ** 3
{{ 2 | pow 3 }}
```
## toYaml
将参数转换成yaml字符串. 参数为左移空格数, 值为字符串
```yaml
{{ .yaml_variable | toYaml }}
```
## ipInCIDR
获取IP范围(cidr)内特定下标的IP地址
```yaml
{{ .cidr_variable | ipInCIDR 1 }}
```
