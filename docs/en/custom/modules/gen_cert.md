# gen_cert Module

Validate or generate certificate files.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| root_key | CA private key path | string | No | - |
| root_cert | CA certificate path | string | No | - |
| date | Certificate validity period | string | No | 1y |
| policy | Generation policy: `Always`, `IfNotPresent`, `None` | string | No | IfNotPresent |
| sans | Subject Alternative Names (IP/DNS list) | string array | No | - |
| cn | Common Name | string | Yes | - |
| out_key | Output private key path | string | Yes | - |
| out_cert | Output certificate path | string | Yes | - |

**policy**:

- **Always**: Always regenerate and overwrite `out_key` / `out_cert`.
- **IfNotPresent**: Generate if not exist; if exists, validate, regenerate if validation fails.
- **None**: Only validate existing files, do not generate; if not exist, do nothing.

## Examples

**1. Generate self-signed CA**

Leave `root_key` and `root_cert` empty when generating CA.

```yaml
- name: Generate root CA
  gen_cert:
    cn: root
    date: 87600h
    policy: IfNotPresent
    out_key: /tmp/pki/root.key
    out_cert: /tmp/pki/root.crt
```

**2. Validate or sign certificate**

Non-CA certificates require existing `root_key` and `root_cert`.

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

`when` uses [template syntax](../101-syntax.md).
