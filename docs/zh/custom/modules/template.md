# template 模块

将模板文件用 [模板语法](../101-syntax.md) 渲染后，复制到连接的目标主机。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| src | 模板文件或目录路径 | 字符串 | 是（与 content 二选一） | - |
| dest | 目标主机上的路径 | 字符串 | 是 | - |

- 相对路径相对于当前 task 对应的 `templates` 目录；任务路径由 task 的 annotation `kubesphere.io/rel-path` 指定。
- 也可使用绝对路径指向本地模板文件。

## 示例

**1. 复制相对路径模板**

```yaml
- name: template relative path
  template:
    src: a.yaml
    dest: /tmp/b.yaml
```

**2. 复制绝对路径模板**

```yaml
- name: template absolute path
  template:
    src: /tmp/a.yaml
    dest: /tmp/b.yaml
```

**3. 复制模板目录**

将目录下所有模板渲染后复制到目标主机。

```yaml
- name: template dir
  template:
    src: /tmp/tpl
    dest: /tmp
```
