/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package modules

import (
	"context"
	"encoding/json"
	"fmt"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Prometheus module uses a dedicated Prometheus connector to query Prometheus data.
This module allows users to run PromQL queries against Prometheus servers defined in the inventory.

Configuration:
Users should define prometheus hosts in the inventory and specify connection information in the connector configuration:

prometheus:
  connector:
    type: prometheus
    host: http://prometheus-server:9090    # The URL of the Prometheus server
    username: admin                        # optional: basic auth username
    password: password                     # optional: basic auth password
    token: my-token                        # optional: Bearer token for authentication
    timeout: 15s                           # optional: request timeout (default: 10s)
    headers:                               # optional: custom HTTP headers
      X-Custom-Header: custom-value

Usage Examples in Playbook Tasks:
1. Basic query:
   ```yaml
   - name: Get Prometheus metrics
     prometheus:
       query: up
     register: prometheus_result
   ```

2. Query with formatting options:
   ```yaml
   - name: Get CPU idle time
     prometheus:
       query: sum(rate(node_cpu_seconds_total{mode='idle'}[5m]))
       format: value
     register: cpu_idle
   ```

3. Query with time parameter:
   ```yaml
   - name: Get historical goroutines count
     prometheus:
       query: go_goroutines
       time: 2023-01-01T12:00:00Z
     register: goroutines
   ```

4. Get Prometheus server info:
   ```yaml
   - name: Get Prometheus server information
     prometheus:
       info: true
     register: prometheus_info
   ```

Format Options:
- raw: Return raw JSON response from Prometheus
- value: Extract single scalar/vector value (when possible)
- table: Format vector results as a table with metric, value, and timestamp columns
*/

// ModulePrometheus handles the "prometheus" module, using prometheus connector to execute PromQL queries
func ModulePrometheus(ctx context.Context, options ExecOptions) (string, string) {
	// Get the hostname to execute on
	host := options.Host
	if host == "" {
		return "", "host name is empty, please specify a prometheus host"
	}

	// Get or create Prometheus connector
	conn, err := getConnector(ctx, host, options.Variable)
	if err != nil {
		return "", fmt.Sprintf("failed to get prometheus connector: %v", err)
	}

	// Initialize connection
	if err := conn.Init(ctx); err != nil {
		return "", fmt.Sprintf("failed to initialize prometheus connector: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(ctx); closeErr != nil {
			// Just log the error, don't return it as we've already executed the main command
			fmt.Printf("Warning: failed to close prometheus connection: %v\n", closeErr)
		}
	}()

	// Get all variables
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	args := variable.Extension2Variables(options.Args)

	// Check if user wants server info instead of running a query
	infoOnly, _ := variable.BoolVar(ha, args, "info")
	if infoOnly != nil && *infoOnly {
		// Try to cast connector to PrometheusConnector
		pc, ok := conn.(*connector.PrometheusConnector)
		if !ok {
			return "", "connector is not a prometheus connector, cannot get server info"
		}

		// Get server info
		info, err := pc.GetServerInfo(ctx)
		if err != nil {
			return "", fmt.Sprintf("failed to get prometheus server info: %v", err)
		}

		// Marshal the info to JSON
		infoBytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return "", fmt.Sprintf("failed to marshal server info: %v", err)
		}

		return string(infoBytes), ""
	}

	// Get query parameters
	query, err := variable.StringVar(ha, args, "query")
	if err != nil {
		return "", "failed to get prometheus query. Please provide a query parameter."
	}

	// Get optional parameters
	format, _ := variable.StringVar(ha, args, "format")
	timeParam, _ := variable.StringVar(ha, args, "time")

	// Build command (include all parameters in JSON format)
	cmdMap := map[string]string{
		"query": query,
	}

	if format != "" {
		cmdMap["format"] = format
	}

	if timeParam != "" {
		cmdMap["time"] = timeParam
	}

	cmdBytes, err := json.Marshal(cmdMap)
	if err != nil {
		return "", fmt.Sprintf("failed to marshal query params: %v", err)
	}

	// Execute query
	result, err := conn.ExecuteCommand(ctx, string(cmdBytes))
	if err != nil {
		return "", fmt.Sprintf("failed to execute prometheus query: %v", err)
	}

	return string(result), ""
}

func init() {
	// Register prometheus module
	utilruntime.Must(RegisterModule("prometheus", ModulePrometheus))
}
