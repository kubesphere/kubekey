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

func ModuleAssert(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.Variable.Get(variable.GetAllVariable(options.Host))
	if err != nil {
		klog.V(4).ErrorS(err, "failed to get host variable", "hostname", options.Host)
		return "", err.Error()
	}

	args := variable.Extension2Variables(options.Args)
	thatParam, err := variable.StringSliceVar(ha.(map[string]any), args, "that")
	if err != nil {
		return "", "\"that\" should be []string or string"
	}

	ok, err := tmpl.ParseBool(ha.(map[string]any), thatParam)
	if err != nil {
		return "", err.Error()
	}

	if ok {
		if successMsgParam, err := variable.StringVar(ha.(map[string]any), args, "success_msg"); err == nil {
			if r, err := tmpl.ParseString(ha.(map[string]any), successMsgParam); err != nil {
				return "", err.Error()
			} else {
				return r, ""
			}
		}
		return "True", ""
	} else {
		if failMsgParam, err := variable.StringVar(ha.(map[string]any), args, "fail_msg"); err == nil {
			if r, err := tmpl.ParseString(ha.(map[string]any), failMsgParam); err != nil {
				return "", err.Error()
			} else {
				return "False", r
			}
		}
		if msgParam, err := variable.StringVar(ha.(map[string]any), args, "msg"); err == nil {
			if r, err := tmpl.ParseString(ha.(map[string]any), msgParam); err != nil {
				return "", err.Error()
			} else {
				return "False", r
			}
		}
		return "False", "False"
	}
}
