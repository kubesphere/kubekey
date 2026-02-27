# http_get_file 模块

从http文件服务拉取文件到本地。

## 参数

| 参数   | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| url  | http文件服务上的文件路径 | 字符串 | 是 | - |
| dest | 本地保存路径 | 字符串 | 是 | - |
| username | Basic 认证用户名 | 字符串 | 否 | - |
| password | Basic 认证密码 | 字符串 | 否 | - |
| token | Bearer Token | 字符串 | 否 | - |
| timeout | 请求超时 | 字符串 | 否 | 10s |
| headers | 自定义 HTTP 头 | map | 否 | - |

## 示例

**1. 拉取文件**

```yaml
- name: http fetch file
  when: '{{ fileExist .dest | not }}'
  http_get_file:
    url: "{{ tpl .artifact_url . }}"     # http文件服务上的文件路径
    dest: "{{ .dest }}"                   # 本地保存路径
    username: admin                       # 可选：Basic 认证用户名
    password: password                    # 可选：Basic 认证密码
    token: my-token                       # 可选：Bearer Token
    timeout: 10s                          # 可选：请求超时，默认 10s
    headers:                              # 可选：自定义 HTTP 头
      X-Custom-Header: custom-value
  vars:
    version: v4.0.3
    artifact_url: "http://localhost/{{\"{{\"}} .version {{\"}}\"}}/test.tar.gz"
    dest: /tmp/{{base .artifact_url}}
```