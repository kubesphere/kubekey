# gen_cert Module

The gen_cert module allows users to validate or generate certificate files.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| root_key | Path to the CA certificate key | string | No | - |
| root_cert | Path to the CA certificate | string | No | - |
| date | Certificate expiration duration | string | No | 1y |
| policy | Certificate generation policy (Always, IfNotPresent, None) | string | No | IfNotPresent |
| sans | Subject Alternative Names. Allowed IPs and DNS | string | No | - |
| cn | Common Name | string | Yes | - |
| out_key | Path to generate the certificate key | string | Yes | - |
| out_cert | Path to generate the certificate | string | Yes | - |

Certificate generation policy:

- **Always**: Always regenerate the certificate and overwrite existing files, regardless of whether `out_key` and `out_cert` exist.
- **IfNotPresent**: Generate a new certificate only if `out_key` and `out_cert` do not exist; if files exist, validate them first and regenerate only if validation fails.
- **None**: If `out_key` and `out_cert` exist, only validate them without generating or overwriting; if files do not exist, no new certificate will be generated.

This policy allows flexible control of certificate generation and validation to meet different scenarios.

## Usage Examples

1. Generate a self-signed CA certificate
When generating a CA certificate, `root_key` and `root_cert` should be empty.
```yaml
- name: Generate root CA file
  gen_cert:
    cn: root
    date: 87600h
    policy: IfNotPresent
    out_key: /tmp/pki/root.key
    out_cert: /tmp/pki/root.crt
```

2. Validate or issue a certificate
For non-CA certificates, `root_key` and `root_cert` should point to an existing CA certificate.
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