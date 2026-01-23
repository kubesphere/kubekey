# copy 模块

将文件、目录或内联内容复制到连接的目标主机。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| src | 源文件或目录路径 | 字符串 | 否（`content` 为空时必填） | - |
| content | 内联内容，直接写入目标 | 字符串 | 否（`src` 为空时必填） | - |
| dest | 目标主机上的路径 | 字符串 | 是 | - |

- 相对路径相对于当前 task 对应的 `files` 目录；任务路径由 task 的 annotation `kubesphere.io/rel-path` 指定。
- 也可使用绝对路径指向本地文件或目录。

## 示例

**1. 复制相对路径文件**

```yaml
- name: copy relative path
  copy:
    src: a.yaml
    dest: /tmp/b.yaml
```

**2. 复制绝对路径文件**

```yaml
- name: copy absolute path
  copy:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

**3. 复制目录**

将该目录下所有文件与子目录复制到目标主机。

```yaml
- name: copy dir
  copy:
    src: /tmp
    dest: /tmp
```

**4. 复制内联内容**

```yaml
- name: copy content
  copy:
    content: hello
    dest: /tmp/b.txt
```
