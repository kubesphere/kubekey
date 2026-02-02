# prometheus 模块

通过 Prometheus 连接器查询 Prometheus 的指标数据，支持 PromQL 查询、结果格式化及服务器信息获取。

## 前置配置

在 [inventory](../201-variable.md#节点清单) 中配置 Prometheus 连接信息，例如：

```yaml
prometheus:
  connector:
    type: prometheus
    host: http://prometheus-server:9090   # Prometheus 地址
    username: admin                       # 可选：Basic 认证用户名
    password: password                    # 可选：Basic 认证密码
    token: my-token                       # 可选：Bearer Token
    timeout: 15s                          # 可选：请求超时，默认 10s
    headers:                              # 可选：自定义 HTTP 头
      X-Custom-Header: custom-value
```

## 参数

| 参数 | 说明 | 类型 | 必填 | 默认值 |
|------|------|------|------|--------|
| query | PromQL 查询 | 字符串 | 是（未使用 info 时） | - |
| format | 结果格式：`raw`、`value`、`table` | 字符串 | 否 | raw |
| time | 查询时间点（RFC3339 或 Unix 时间戳） | 字符串 | 否 | 当前时间 |

**format**：

- **raw**：原始 JSON
- **value**：单个标量/向量值（如适用）
- **table**：表格，含 metric、value、timestamp 等列

## 示例

**1. 基础查询**

```yaml
- name: query Prometheus
  prometheus:
    query: up
  register: prometheus_result
```

**2. 指定格式化**

```yaml
- name: CPU idle
  prometheus:
    query: sum(rate(node_cpu_seconds_total{mode='idle'}[5m]))
    format: value
  register: cpu_idle
```

**3. 指定时间**

```yaml
- name: goroutines at time
  prometheus:
    query: go_goroutines
    time: 2023-01-01T12:00:00Z
  register: goroutines
```

**4. 表格格式**

```yaml
- name: node CPU as table
  prometheus:
    query: node_cpu_seconds_total{mode="idle"}
    format: table
  register: cpu_table
```

## 注意

- 执行查询时 `query` 必填。
- `time` 需为 RFC3339（如 `2023-01-01T12:00:00Z`）或 Unix 时间戳。
- `format: table` 仅适用于向量结果。
- 生产环境建议使用 HTTPS。
