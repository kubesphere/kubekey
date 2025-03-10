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

	"gopkg.in/yaml.v2"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ModuleSetFact deal "set_fact" module
func ModuleSetFact(_ context.Context, options ExecOptions) (string, string) {
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}
	// get host variable
	args := variable.Extension2Variables(options.Args)
	for k, v := range args {
		switch val := v.(type) {
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			args[k] = val
		case string:
			sv, err := tmpl.Parse(ha, val)
			if err != nil {
				return "", fmt.Sprintf("parse %q error: %v", k, err)
			}
			var ssvResult any
			if json.Valid(sv) {
				_ = json.Unmarshal(sv, &ssvResult)
			} else {
				_ = yaml.Unmarshal(sv, &ssvResult)
			}
			args[k] = ssvResult
		default:
			return "", fmt.Sprintf("only support bool, int, float64, string value for %q.", k)
		}
	}
	if err := options.Variable.Merge(variable.MergeAllRuntimeVariable(args, options.Host)); err != nil {
		return "", fmt.Sprintf("set_fact error: %v", err)
	}

	return StdoutSuccess, ""
}
