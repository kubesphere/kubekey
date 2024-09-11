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
	"fmt"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
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
	// var is defined. return the value of var
	if varParam, err := variable.StringVar(ha, args, "var"); err == nil {
		result, err := tmpl.ParseString(ha, fmt.Sprintf("{{ %s }}", varParam))
		if err != nil {
			return "", fmt.Sprintf("failed to parse var: %v", err)
		}

		return result, ""
	}
	// msg is defined. return the actual msg
	if msgParam, err := variable.StringVar(ha, args, "msg"); err == nil {
		return msgParam, ""
	}

	return "", "unknown args for debug. only support var or msg"
}
