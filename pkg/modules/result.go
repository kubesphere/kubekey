package modules

import (
	"context"

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
func ModuleResult(ctx context.Context, options ExecOptions) (string, string, error) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}
	arg, err := variable.Extension2String(ha, options.Args)
	if err != nil {
		return StdoutFailed, StderrParseArgument, err
	}
	var result any
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(arg, &result); err != nil {
		return StdoutFailed, "failed to unmarshal YAML", err
	}

	if err := options.Variable.Merge(variable.MergeResultVariable(result)); err != nil {
		return StdoutFailed, "failed to merge result variable", err
	}

	return StdoutSuccess, "", nil
}

func init() {
	utilruntime.Must(RegisterModule("result", ModuleResult))
}
