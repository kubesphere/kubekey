# copy 模块

copy模块允许用户复制文件或文件夹到连接的目标主机。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| src | 原始文件或文件夹路径 | 字符串 | 否（content空时必填） | - |
| content | 原始文件或文件夹内容 | 字符串 | 否（src空时必填） | - |
| dest | 复制到目标主机的文件目录 | 字符串 | 是 | - |

## 使用示例

1. 复制相对路径文件到目标主机
相对路径是当前任务对应的`files`目录中。当前任务路径由task的annotations `kubesphere.io/rel-path` 指定
```yaml
- name: copy relative path
  copy:
    src: a.yaml
    dest: /tmp/b.yaml
```

2. 复制绝对路径文件到目标主机
本地绝对路径的文件
```yaml
- name: copy absolute path
  copy:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

3. 复制目录到目标主机
复制该目录下所有的文件和目录到目标主机
```yaml
- name: copy dir
  copy:
    src: /tmp
    dest: /tmp
```

4. 复制文件内容到目标主机
```yaml
- name: copy content
  copy:
    content: hello
    dest: /tmp/b.txt
```