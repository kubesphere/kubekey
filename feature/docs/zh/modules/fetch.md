# fetch 模块

fetch模块允许用户从远程主机上拉取文件到本地。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| src | 远程主机上要拉取的文件路径 | 字符串 | 是 | - |
| dest | 拉取到本地的文件路径 | 字符串 | 是 | - |

## 使用示例

1. 拉取文件
```yaml
- name: fetch file
  fetch:
    src: /tmp/src.yaml
    dest: /tmp/dest.yaml
```