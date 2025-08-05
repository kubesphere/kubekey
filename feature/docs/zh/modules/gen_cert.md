# gen_cert 模块

gen_cert模块允许用户校验或生成证书文件。

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| root_key | CA证书的key路径 | 字符串 | 否 | - |
| root_cert | CA证书路径 | 字符串 | 否 | - |
| date | 证书过期时间 | 字符串 | 否 | 1y |
| policy | 证书生成策略（Always, IfNotPresent） | 字符串 | 否 | IfNotPresent |
| sans | Subject Alternative Names. 允许的IP和DNS | 字符串 | 否 | - |
| cn | Common Name | 字符串 | 是 | - |
| out_key | 生成的证书key路径 | 字符串 | 是 | - |
| out_cert | 生成的证书路径 | 字符串 | 是 | - |

证书生成策略:
Always: 无论`out_key`和`out_cert`指向的证书路径是否存在，都会生成新的证书进行覆盖。
IfNotPresent: 当`out_key`和`out_cert`指向的证书路径存在时，只对该证书进行校验，并不会生成新的证书。

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