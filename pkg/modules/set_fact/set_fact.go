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

package set_fact

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The SetFact module allows setting variables during playbook execution.
This module enables users to define and update variables that can be used in subsequent tasks.

Configuration:
Users can specify key-value pairs to set as variables:

set_fact:
  key1: value1    # required: variable name and value
  key2: value2    # optional: additional variables

Usage Examples in Playbook Tasks:
1. Set single variable:
   ```yaml
   - name: Set version variable
     set_fact:
       app_version: "1.0.0"
     register: version_result
   ```

2. Set multiple variables:
   ```yaml
   - name: Set configuration variables
     set_fact:
       db_host: "localhost"
       db_port: 5432
     register: config_vars
   ```

Return Values:
- On success: Returns "Success" in stdout
- On failure: Returns error message in stderr
*/

// ModuleSetFact handles the "set_fact" module, setting variables during playbook execution
func ModuleSetFact(_ context.Context, opts internal.ExecOptions) (string, string, error) {
	var node yaml.Node
	// Unmarshal the YAML document into a root node.
	if err := yaml.Unmarshal(opts.Args.Raw, &node); err != nil {
		return internal.StdoutFailed, "failed to unmarshal YAML", err
	}
	if err := opts.Merge(variable.MergeRuntimeVariable([]yaml.Node{node}, opts.Host)); err != nil {
		return internal.StdoutFailed, "failed to merge set_fact variable", err
	}

	return internal.StdoutSuccess, "", nil
}
