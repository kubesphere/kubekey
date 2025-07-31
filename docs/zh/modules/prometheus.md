# Prometheus 模块

Prometheus模块允许用户查询Prometheus服务器中的指标数据。它通过专用的Prometheus连接器实现，支持运行PromQL查询、格式化结果以及获取服务器信息。

## 配置

要使用Prometheus模块，需要在清单（inventory）中定义Prometheus主机并指定连接信息：

```yaml
prometheus:
  connector:
    type: prometheus
    host: http://prometheus-server:9090    # Prometheus服务器的URL
    username: admin                        # 可选: 基本认证用户名
    password: password                     # 可选: 基本认证密码
    token: my-token                        # 可选: Bearer令牌认证
    timeout: 15s                           # 可选: 请求超时（默认10秒）
    headers:                               # 可选: 自定义HTTP头
      X-Custom-Header: custom-value
```

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|-------|
| query | PromQL查询语句 | 字符串 | 是（除非使用info参数） | - |
| format | 结果格式化选项：raw、value、table | 字符串 | 否 | raw |
| time | 查询的时间点（RFC3339格式或Unix时间戳） | 字符串 | 否 | 当前时间 |

## 输出

模块返回查询结果或服务器信息，格式取决于指定的format参数：

- **raw**: 返回原始JSON响应
- **value**: 提取单个标量/向量值（如可能）
- **table**: 将向量结果格式化为表格，包含指标、值和时间戳列

## 使用示例

1. 基本查询：
```yaml
- name: 获取Prometheus指标
  prometheus:
    query: up
  register: prometheus_result
```

2. 格式化选项：
```yaml
- name: 获取CPU空闲时间
  prometheus:
    query: sum(rate(node_cpu_seconds_total{mode='idle'}[5m]))
    format: value
  register: cpu_idle
```

3. 指定时间参数：
```yaml
- name: 获取历史Goroutines数量
  prometheus:
    query: go_goroutines
    time: 2023-01-01T12:00:00Z
  register: goroutines
```

4. 获取Prometheus服务器信息：
```yaml
- name: 获取Prometheus服务器信息
  fetch:
    src: api/v1/status/buildinfo
    dest: info.json
```

5. 使用表格格式化结果：
```yaml
- name: 获取节点CPU使用率并格式化为表格
  prometheus:
    query: node_cpu_seconds_total{mode="idle"}
    format: table
  register: cpu_table
```

## 注意事项

1. 如果需要执行查询，`query`参数是必需的
2. 时间参数需要符合RFC3339格式（如：2023-01-01T12:00:00Z）或Unix时间戳格式
3. 表格格式化仅适用于向量类型的结果，其他类型的结果会返回错误
4. 为了确保安全，推荐使用HTTPS连接到Prometheus服务器
