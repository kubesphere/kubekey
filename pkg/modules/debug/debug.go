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

package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Debug module provides debugging capabilities by printing variable values and messages.
This module allows users to inspect variable values and debug playbook execution.

Configuration:
Users can specify either a message or a variable path to debug.

debug:
  msg: "message"                    # optional: direct message to print
  msg: "value is {{ .var }}"        # optional: template syntax with filters
  var: "{{ .var }}"                 # optional: template syntax to get variable
  var: ".var"                       # optional: simple variable path

Usage Examples in Playbook Tasks:
1. Print direct message:
   ```yaml
   - name: Debug message
     debug:
       msg: "Starting deployment"
     register: debug_result
   ```

2. Print message with template:
   ```yaml
   - name: Debug with template
     debug:
       msg: "Version is {{ .config.version | default \"unknown\" }}"
     register: debug_result
   ```

3. Print variable value:
   ```yaml
   - name: Debug variable
     debug:
       var: "{{ .config.version }}"
     register: var_debug
   ```

4. Print variable with simple path:
   ```yaml
   - name: Debug variable simple
     debug:
       var: ".config.version"
     register: var_debug
   ```

Return Values:
- On success: Returns formatted message/variable value in stdout
- On failure: Returns error message in stderr
*/

// ModuleDebug handles the "debug" module, printing debug information
func ModuleDebug(_ context.Context, opts internal.ExecOptions) (string, string, error) {
	// get host variable
	ha, err := opts.GetAllVariables()
	if err != nil {
		return internal.StdoutFailed, internal.StderrGetHostVariable, err
	}
	args := variable.Extension2Variables(opts.Args)

	// Handle "var" field - for getting variable values
	if v, ok := args["var"]; ok {
		return handleVarField(v, ha, opts.LogOutput)
	}

	// Handle "msg" field - for printing messages with template support
	if v, ok := args["msg"]; ok {
		return handleMsgField(v, ha, opts.LogOutput)
	}

	return internal.StdoutFailed, internal.StderrUnsupportArgs, errors.New("either \"msg\" or \"var\" must be specified")
}

// handleVarField handles the "var" field for variable debugging
// Supports both template syntax "{{ .var }}" and simple path ".var"
func handleVarField(v any, ha map[string]any, output io.Writer) (string, string, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		val := rv.String()
		// Check if it's template syntax
		if kkprojectv1.IsTmplSyntax(val) {
			val = kkprojectv1.TrimTmplSyntax(val)
		}
		// Ensure it starts with "."
		if !strings.HasPrefix(val, ".") {
			return internal.StdoutFailed, internal.StderrUnsupportArgs, errors.New("variable path must start with '.'")
		}
		// Remove leading "." and split path
		path := strings.TrimPrefix(val, ".")
		if path == "" {
			return internal.StdoutFailed, internal.StderrUnsupportArgs, errors.New("variable path cannot be empty")
		}
		data, err := variable.PrintVar(ha, strings.Split(path, ".")...)
		if err != nil {
			return internal.StdoutFailed, internal.StderrParseArgument, err
		}
		return formatOutput(data, output), "", nil
	default:
		// For non-string types, pass directly to formatOutput for pretty printing
		return formatOutput(rv.Interface(), output), "", nil
	}
}

// handleMsgField handles the "msg" field for message debugging
// Supports template syntax with filters like "a is {{ .var | default 'b' }}"
func handleMsgField(v any, ha map[string]any, output io.Writer) (string, string, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		msg := rv.String()
		// If it contains template syntax, parse it
		if kkprojectv1.IsTmplSyntax(msg) {
			parsed, err := tmpl.Parse(ha, msg)
			if err != nil {
				return internal.StdoutFailed, internal.StderrParseArgument, err
			}
			return formatOutput(string(parsed), output), "", nil
		}
		// Direct message without template
		return formatOutput(msg, output), "", nil
	default:
		// For non-string types, pass directly to formatOutput for pretty printing
		return formatOutput(rv.Interface(), output), "", nil
	}
}

// formatOutput formats data as pretty JSON and logs it with DEBUG prefix if output is provided
// Returns the formatted string
func formatOutput(data any, output io.Writer) string {
	var msg string
	// Handle string data directly without JSON marshaling
	switch v := data.(type) {
	case string:
		msg = v
	case []byte:
		msg = string(v)
	default:
		prettyJSON, err := json.MarshalIndent(v, "", "  ")
		if err == nil {
			msg = string(prettyJSON)
		}
	}
	if output != nil {
		_, _ = fmt.Fprintln(output, "DEBUG: \n"+msg) // Ignore error in test context
	}
	return msg
}
