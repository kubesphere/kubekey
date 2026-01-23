# gen_cert 模块

校验或生成证书文件。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| root_key | CA 私钥路径 | 字符串 | 否 | - |
| root_cert | CA 证书路径 | 字符串 | 否 | - |
| date | 证书有效期 | 字符串 | 否 | 1y |
| policy | 生成策略：`Always`、`IfNotPresent`、`None` | 字符串 | 否 | IfNotPresent |
| sans | Subject Alternative Names（IP/DNS 列表） | 字符串数组 | 否 | - |
| cn | Common Name | 字符串 | 是 | - |
| out_key | 输出私钥路径 | 字符串 | 是 | - |
| out_cert | 输出证书路径 | 字符串 | 是 | - |

**policy**：

- **Always**：始终重新生成并覆盖 `out_key` / `out_cert`。
- **IfNotPresent**：不存在则生成；已存在则校验，不通过再重新生成。
- **None**：仅校验已存在文件，不生成；不存在则不做任何操作。

## 示例

**1. 生成自签名 CA**

生成 CA 时 `root_key`、`root_cert` 留空。

```yaml
- name: Generate root CA
  gen_cert:
    cn: root
    date: 87600h
    policy: IfNotPresent
    out_key: /tmp/pki/root.key
    out_cert: /tmp/pki/root.crt
```

**2. 校验或签发证书**

非 CA 证书需指定已有的 `root_key`、`root_cert`。

```yaml
- name: Generate server cert
  gen_cert:
    root_key: /tmp/pki/root.key
    root_cert: /tmp/pki/root.crt
    cn: server
    sans:
      - 127.0.0.1
      - localhost
    date: 87600h
    policy: IfNotPresent
    out_key: /tmp/pki/server.key
    out_cert: /tmp/pki/server.crt
  when: .groups.image_registry | default list | empty | not
```

`when` 使用 [模板语法](../101-syntax.md)。
