# gen_cert 模块

gen_cert模块允许用户校验或生成证书文件。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| root_key | CA证书的key路径 | 字符串 | 否 | - |
| root_cert | CA证书路径 | 字符串 | 否 | - |
| date | 证书过期时间 | 字符串 | 否 | 1y |
| policy | 证书生成策略（Always, IfNotPresent, None） | 字符串 | 否 | IfNotPresent |
| sans | Subject Alternative Names. 允许的IP和DNS | 字符串 | 否 | - |
| cn | Common Name | 字符串 | 是 | - |
| out_key | 生成的证书key路径 | 字符串 | 是 | - |
| out_cert | 生成的证书路径 | 字符串 | 是 | - |

证书生成策略说明：

- **Always**：无论`out_key`和`out_cert`指定的证书文件是否已存在，始终重新生成证书并覆盖原有文件。
- **IfNotPresent**：仅当`out_key`和`out_cert`指定的证书文件不存在时才生成新证书；如果文件已存在，则先对现有证书进行校验，校验不通过时才会重新生成证书。
- **None**：如果`out_key`和`out_cert`指定的证书文件已存在，仅对其进行校验，不会生成或覆盖证书文件；若文件不存在则不会生成新证书。

该策略可灵活控制证书的生成和校验行为，满足不同场景下的需求。

## 使用示例

1. 生成自签名的CA证书
生成CA证书时，`root_key`和`root_cert`需为空
```yaml
- name: Generate root ca file
  gen_cert:
    cn: root
    date: 87600h
    policy: IfNotPresent
    out_key: /tmp/pki/root.key
    out_cert: /tmp/pki/root.crt
```

2. 校验或签发证书
对于非CA证书。`root_key`和`root_cert`需指向已存在CA证书。
```yaml
- name: Generate registry image cert file
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