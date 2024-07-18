# 语法
语法遵循Django-syntax规范.采用[pongo2](https://github.com/flosch/pongo2)实现, 并pongo2的关键字进行了扩展
# 自定义关键字
## defined
判断某个参数是否在[variable](201-variable.md)中定义. 值为bool类型
```yaml
{{ variable | defined }}
```
## version
比较版本大小. 参数为比较标准, 值为bool类型
```yaml
# version_variable>v1.0.0
{{ version_variable | version:'>v1.0.0' }}
# version_variable>=v1.0.0
{{ version_variable | version:'>=v1.0.0' }}
# version_variable==v1.0.0
{{ version_variable | version:'==v1.0.0' }}
# version_variable<=v1.0.0
{{ version_variable | version:'<=v1.0.0' }}
# version_variable<v1.0.0
{{ version_variable | version:'<v1.0.0' }}
```
## pow
幂运算. 参数为幂, 值为浮点类型
```yaml
# 2的3次方
{{ 2 | pow:'3' }}
```
## match
正则匹配. 参数为正则表达式, 值为bool类型
```yaml
# 判断string_variable是否匹配正则表达式'*+'
{{ string_variable | match:'*+' }}
```
## to_json
将参数转换成json字符串. 值为字符串
```yaml
{{ variable | to_json }}
```
## to_yaml
将参数转换成yaml字符串. 参数为左移空格数, 值为字符串
```yaml
{{ variable | to_json }}
```
## ip_range
将ip范围(cidr)转换成可用的ip数组, 值为字符串数组
```yaml
{{ string_variable | ip_range }}
{{ string_variable | ip_range:1 }}
{{ string_variable | ip_range:"-1" }}
{{ string_variable | ip_range:"1:2" }}
```
无参数: 返回一整个可用ip列表
参数为数字: 返回对应下标的ip(负数从最大的开始, 即-1表示返回ip列表最后一个ip)
参数为字符串形数字: 返回对应下标的ip(负数从最大的开始, 即-1表示返回ip列表最后一个ip)
参数为"数字:数字": 返回对应范围的ip列表(负数从最大的开始, 即-1表示返回ip列表最后一个ip)

## get
获取map类型的variable中某个key值对应的value, 参数为key, 值为任意类型
```yaml
{{ map_variable | get:'key' }}
```
