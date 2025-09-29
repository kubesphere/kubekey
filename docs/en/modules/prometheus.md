# Prometheus Module

The Prometheus module allows users to query metric data from a Prometheus server. It uses a dedicated Prometheus connector and supports running PromQL queries, formatting results, and fetching server information.

## Configuration

To use the Prometheus module, define Prometheus hosts and connection info in the inventory:

```yaml
prometheus:
  connector:
    type: prometheus
    host: http://prometheus-server:9090    # URL of the Prometheus server
    username: admin                        # Optional: basic auth username
    password: password                     # Optional: basic auth password
    token: my-token                        # Optional: Bearer token
    timeout: 15s                           # Optional: request timeout (default 10s)
    headers:                               # Optional: custom HTTP headers
      X-Custom-Header: custom-value
```

## Parameters

| Parameter | Description | Type | Required | Default |
|-----------|------------|------|---------|---------|
| query     | PromQL query statement | string | Yes (unless using info) | - |
| format    | Result format: raw, value, table | string | No | raw |
| time      | Query time (RFC3339 or Unix timestamp) | string | No | current time |

## Output

The module returns query results or server information depending on the specified format:

- **raw**: returns the original JSON response
- **value**: extracts a single scalar/vector value if possible
- **table**: formats vector results as a table with columns for metric, value, and timestamp

## Usage Examples

1. Basic query:
```yaml
- name: Get Prometheus metrics
  prometheus:
    query: up
  register: prometheus_result
```

2. With format option:
```yaml
- name: Get CPU idle time
  prometheus:
    query: sum(rate(node_cpu_seconds_total{mode='idle'}[5m]))
    format: value
  register: cpu_idle
```

3. Specify time parameter:
```yaml
- name: Get historical Goroutines count
  prometheus:
    query: go_goroutines
    time: 2023-01-01T12:00:00Z
  register: goroutines
```

4. Fetch Prometheus server information:
```yaml
- name: Fetch Prometheus server info
  fetch:
    src: api/v1/status/buildinfo
    dest: info.json
```

5. Format results as table:
```yaml
- name: Get node CPU usage and format as table
  prometheus:
    query: node_cpu_seconds_total{mode="idle"}
    format: table
  register: cpu_table
```

## Notes

1. The `query` parameter is required when executing queries
2. Time must be in RFC3339 format (e.g., 2023-01-01T12:00:00Z) or Unix timestamp
3. Table formatting only applies to vector results; other types will return an error
4. For security, HTTPS connections to Prometheus are recommended
