# debug 模块

debug模块允许用户打印变量

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| msg | 原始文件或文件夹路径 | 字符串 | 是 | - |

## 使用示例

1. 打印字符串
```yaml
- name: debug string
  debug:
    msg: I'm {{ .name }}
```
当变量`name`为kubekey时，打印内容如下
```txt
DEBUG: 
I'm kubekey
```

2. 打印map
```yaml
- name: debug map
  debug:
    msg: >-
      {{ .product }}
```
当变量`product`为map时，比如`{"name":"kubekey"}`，打印内容如下
```txt
DEBUG: 
{
    "name": "kubekey"
}
```

3. 打印数组
```yaml
- name: debug array
  debug:
    msg: >-
      {{ .version }}
```
当变量`version`为array时，比如`["1","2"]`，打印内容如下
```txt
DEBUG: 
[
    "1",
    "2"
]
```