package modules

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Result module allows setting result variables during playbook execution.
This module enables users to define and update result variables that can be accessed
by subsequent tasks in the same playbook.

Configuration:
Users can specify key-value pairs to set as result variables:

result:
  key1: value1    # required: result variable name and value
  key2: value2    # optional: additional result variables

Usage Examples in Playbook Tasks:
1. Set single result variable:
   ```yaml
   - name: Set result variable
     result:
       app_version: "1.0.0"
     register: version_result
   ```

2. Set multiple result variables:
   ```yaml
   - name: Set result configuration variables
     result:
       db_host: "localhost"
       db_port: 5432
     register: config_vars
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

// ModuleResult handles the "result" module, setting result variables during playbook execution
func ModuleResult(ctx context.Context, options ExecOptions) (string, string) {
	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(options.Args.Raw, &node); err != nil {
		return "", fmt.Sprintf("failed to unmarshal YAML error: %v", err)
	}

	if err := options.Variable.Merge(variable.MergeResultVariable(node, options.Host)); err != nil {
		return "", fmt.Sprintf("result error: %v", err)
	}

	return StdoutSuccess, ""
}

func init() {
	utilruntime.Must(RegisterModule("result", ModuleResult))
}
