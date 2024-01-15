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

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleSetFact(ctx context.Context, options ExecOptions) (string, string) {
	args := variable.Extension2Variables(options.Args)
	lv, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		klog.ErrorS(err, "failed to get location vars")
		return "", err.Error()
	}

	factVars := variable.VariableData{}
	for k, v := range args {
		switch v.(type) {
		case string:
			factVars[k], err = tmpl.ParseString(lv.(variable.VariableData), v.(string))
			if err != nil {
				klog.ErrorS(err, "template parse error", "input", v)
				return "", err.Error()
			}
		default:
			factVars[k] = v
		}
	}

	if err := options.Variable.Merge(variable.HostMerge{
		HostNames:   []string{options.Host},
		LocationUID: "",
		Data:        factVars,
	}); err != nil {
		klog.ErrorS(err, "merge fact error")
		return "", err.Error()
	}
	return "success", ""
}
