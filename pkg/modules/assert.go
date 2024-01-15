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

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleAssert(ctx context.Context, options ExecOptions) (string, string) {
	args := variable.Extension2Variables(options.Args)
	that := variable.StringSliceVar(args, "that")
	if that == nil {
		st := variable.StringVar(args, "that")
		if st == nil {
			return "", "\"that\" should be []string or string"
		}
		that = []string{*st}
	}
	lg, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		return "", err.Error()
	}
	ok, err := tmpl.ParseBool(lg.(variable.VariableData), that)
	if err != nil {
		return "", err.Error()
	}

	if ok {
		if v := variable.StringVar(args, "success_msg"); v != nil {
			if r, err := tmpl.ParseString(lg.(variable.VariableData), *v); err != nil {
				return "", err.Error()
			} else {
				return r, ""
			}
		}
		return "True", ""
	} else {
		if v := variable.StringVar(args, "fail_msg"); v != nil {
			if r, err := tmpl.ParseString(lg.(variable.VariableData), *v); err != nil {
				return "", err.Error()
			} else {
				return "False", r
			}
		}
		if v := variable.StringVar(args, "msg"); v != nil {
			if r, err := tmpl.ParseString(lg.(variable.VariableData), *v); err != nil {
				return "", err.Error()
			} else {
				return "False", r
			}
		}
		return "False", "False"
	}
}
