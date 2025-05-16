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

	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ModuleDebug deal "debug" module
func ModuleDebug(_ context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}
	args := variable.Extension2Variables(options.Args)
	v := reflect.ValueOf(args["msg"])
	switch v.Kind() {
	case reflect.Invalid:
		return "", "\"msg\" is not found"
	case reflect.String:
		if !kkprojectv1.IsTmplSyntax(v.String()) {
			return FormatOutput([]byte(v.String()), options.LogOutput), ""
		}
		val := kkprojectv1.TrimTmplSyntax(v.String())
		if !strings.HasPrefix(val, ".") {
			return "", "error tmpl value syntax"
		}
		data, err := variable.PrintVar(ha, strings.Split(val, ".")[1:]...)
		if err != nil {
			return "", err.Error()
		}
		return FormatOutput(data, options.LogOutput), ""
	default:
		// do not parse by ctx
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return "", err.Error()
		}
		return FormatOutput(data, options.LogOutput), ""
	}
}

// FormatOutput formats data as pretty JSON and logs it with DEBUG prefix if output is provided
// Returns the formatted string
func FormatOutput(data any, output io.Writer) string {
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
