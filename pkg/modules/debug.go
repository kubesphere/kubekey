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

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleDebug(ctx context.Context, options ExecOptions) (string, string) {
	args := variable.Extension2Variables(options.Args)
	if v := variable.StringVar(args, "var"); v != nil {
		lg, err := options.Variable.Get(variable.LocationVars{
			HostName:    options.Host,
			LocationUID: string(options.Task.UID),
		})
		if err != nil {
			klog.ErrorS(err, "Failed to get location vars")
			return "", err.Error()
		}
		result, err := tmpl.ParseString(lg.(variable.VariableData), fmt.Sprintf("{{ %s }}", *v))
		if err != nil {
			klog.ErrorS(err, "Failed to get var")
			return "", err.Error()
		}
		return result, ""
	}

	if v := variable.StringVar(args, "msg"); v != nil {
		lg, err := options.Variable.Get(variable.LocationVars{
			HostName:    options.Host,
			LocationUID: string(options.Task.UID),
		})
		if err != nil {
			klog.ErrorS(err, "Failed to get location vars")
			return "", err.Error()
		}
		result, err := tmpl.ParseString(lg.(variable.VariableData), *v)
		if err != nil {
			klog.ErrorS(err, "Failed to get var")
			return "", err.Error()
		}
		return result, ""
	}

	return "", "unknown args for debug. only support var or msg"
}
