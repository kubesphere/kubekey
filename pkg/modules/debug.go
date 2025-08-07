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
	"io"
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Debug module provides debugging capabilities by printing variable values and messages.
This module allows users to inspect variable values and debug playbook execution.

Configuration:
Users can specify either a message or a variable path to debug.

debug:
  msg: "message"        # optional: direct message to print
  msg: "{{ .var }}"     # optional: template syntax to print variable value

Usage Examples in Playbook Tasks:
1. Print direct message:
   ```yaml
   - name: Debug message
     debug:
       msg: "Starting deployment"
     register: debug_result
   ```

2. Print variable value:
   ```yaml
   - name: Debug variable
     debug:
       msg: "{{ .config.version }}"
     register: var_debug
   ```

Return Values:
- On success: Returns formatted message/variable value in stdout
- On failure: Returns error message in stderr
*/

// ModuleDebug handles the "debug" module, printing debug information
func ModuleDebug(_ context.Context, options ExecOptions) (string, string, error) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return StdoutFailed, StderrGetHostVariable, err
	}
	args := variable.Extension2Variables(options.Args)
	v := reflect.ValueOf(args["msg"])
	switch v.Kind() {
	case reflect.Invalid:
		return StdoutFailed, StderrUnsupportArgs, errors.New("\"msg\" is not found")
	case reflect.String:
		if !kkprojectv1.IsTmplSyntax(v.String()) {
			return formatOutput([]byte(v.String()), options.LogOutput), "", nil
		}
		val := kkprojectv1.TrimTmplSyntax(v.String())
		if !strings.HasPrefix(val, ".") {
			return StdoutFailed, StderrUnsupportArgs, errors.New("error tmpl value syntax")
		}
		data, err := variable.PrintVar(ha, strings.Split(val, ".")[1:]...)
		if err != nil {
			return StdoutFailed, StderrParseArgument, err
		}
		return formatOutput(data, options.LogOutput), "", nil
	default:
		// do not parse by ctx
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return StdoutFailed, "failed to marshal value to json", err
		}
		return formatOutput(data, options.LogOutput), "", nil
	}
}

// formatOutput formats data as pretty JSON and logs it with DEBUG prefix if output is provided
// Returns the formatted string
func formatOutput(data any, output io.Writer) string {
	var msg string
	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err == nil {
		msg = string(prettyJSON)
	}
	if output != nil {
		_, _ = fmt.Fprintln(output, "DEBUG: \n"+msg) // Ignore error in test context
	}
	return msg
}

func init() {
	utilruntime.Must(RegisterModule("debug", ModuleDebug))
}
