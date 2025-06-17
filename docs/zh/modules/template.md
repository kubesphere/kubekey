# template 模块

template模块允许用户将模板文件解析后复制到连接的目标主机。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| src | 原始文件或文件夹路径 | 字符串 | 否（content空时必填） | - |
| dest | 复制到目标主机的文件目录 | 字符串 | 是 | - |

## 使用示例

1. 复制相对路径文件到目标主机
相对路径是当前任务对应的`templates`目录中。当前任务路径由task的annotations `kubesphere.io/rel-path` 指定
```yaml
- name: copy relative path
  template:
    src: a.yaml
    dest: /tmp/b.yaml
```

2. 复制绝对路径文件到目标主机
本地模板文件绝对路径的文件
```yaml
- name: copy absolute path
  template:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

3. 复制目录到目标主机
将目录下所有的模板文件解析后复制到目标主机
```yaml
- name: copy dir
  template:
    src: /tmp
    dest: /tmp
```

1. 复制文件内容到目标主机
```yaml
- name: copy content
  template:
    content: hello
    dest: /tmp/b.txt
```