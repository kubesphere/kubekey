# prometheus Module

Query Prometheus metric data through the Prometheus connector, supporting PromQL queries, result formatting, and server information retrieval.

## Prerequisite Configuration

Configure Prometheus connection information in the [inventory](../201-variable.md#inventory), for example:

```yaml
prometheus:
  connector:
    type: prometheus
    host: http://prometheus-server:9090   # Prometheus address
    username: admin                       # Optional: Basic auth username
    password: password                    # Optional: Basic auth password
    token: my-token                       # Optional: Bearer Token
    timeout: 15s                          # Optional: request timeout, default 10s
    headers:                              # Optional: custom HTTP headers
      X-Custom-Header: custom-value
```

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|-------------|------|----------|---------|
| query | PromQL query | string | Yes (when not using info) | - |
| format | Result format: `raw`, `value`, `table` | string | No | raw |
| time | Query timestamp (RFC3339 or Unix timestamp) | string | No | Current time |

**format**:

- **raw**: Raw JSON
- **value**: Single scalar/vector value (if applicable)
- **table**: Table format, includes metric, value, timestamp, etc.

## Examples

**1. Basic query**

```yaml
- name: query Prometheus
  prometheus:
    query: up
  register: prometheus_result
```

**2. Specify formatting**

```yaml
- name: CPU idle
  prometheus:
    query: sum(rate(node_cpu_seconds_total{mode='idle'}[5m]))
    format: value
  register: cpu_idle
```

**3. Specify time**

```yaml
- name: goroutines at time
  prometheus:
    query: go_goroutines
    time: 2023-01-01T12:00:00Z
  register: goroutines
```

**4. Table format**

```yaml
- name: node CPU as table
  prometheus:
    query: node_cpu_seconds_total{mode="idle"}
    format: table
  register: cpu_table
```

## Notes

- `query` is required when executing queries.
- `time` must be RFC3339 (e.g., `2023-01-01T12:00:00Z`) or Unix timestamp.
- `format: table` only applies to vector results.
- HTTPS is recommended for production environments.
