# http_get_file Module

Pull files from an HTTP file server to the local machine.

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| url | File path on the HTTP file server | String | Yes | - |
| dest | Local save path | String | Yes | - |
| username | Basic auth username | String | No | - |
| password | Basic auth password | String | No | - |
| token | Bearer Token | String | No | - |
| timeout | Request timeout | String | No | 10s |
| headers | Custom HTTP headers | Map | No | - |

## Examples

**1. Fetch file**

```yaml
- name: http fetch file
  when: '{{ fileExist .dest | not }}'
  http_get_file:
    url: "{{ tpl .artifact_url . }}"     # File path on HTTP file server
    dest: "{{ .dest }}"                   # Local save path
    username: admin                       # Optional: Basic auth username
    password: password                    # Optional: Basic auth password
    token: my-token                       # Optional: Bearer Token
    timeout: 10s                          # Optional: Request timeout, default 10s
    headers:                              # Optional: Custom HTTP headers
      X-Custom-Header: custom-value
  vars:
    version: v4.0.3
    artifact_url: "http://localhost/{{\"{{\"}} .version {{\"}}\"}}/test.tar.gz"
    dest: /tmp/{{base .artifact_url}}
```
