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
	"strings"

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func ModuleCommand(ctx context.Context, options ExecOptions) (string, string) {
	ha, _ := options.Variable.Get(variable.HostVars{HostName: options.Host})
	var conn connector.Connector
	if v := ctx.Value("connector"); v != nil {
		conn = v.(connector.Connector)
	} else {
		conn = connector.NewConnector(options.Host, ha.(variable.VariableData))
	}
	if err := conn.Init(ctx); err != nil {
		klog.Errorf("failed to init connector %v", err)
		return "", err.Error()
	}
	defer conn.Close(ctx)

	// convert command template to string
	arg := variable.Extension2String(options.Args)
	lg, err := options.Variable.Get(variable.LocationVars{
		HostName:    options.Host,
		LocationUID: string(options.Task.UID),
	})
	if err != nil {
		return "", err.Error()
	}
	result, err := tmpl.ParseString(lg.(variable.VariableData), arg)
	if err != nil {
		return "", err.Error()
	}
	// execute command
	var stdout, stderr string
	data, err := conn.ExecuteCommand(ctx, result)
	if err != nil {
		stderr = err.Error()
	}
	if data != nil {
		stdout = strings.TrimSuffix(string(data), "\n")
	}
	return stdout, stderr
}
