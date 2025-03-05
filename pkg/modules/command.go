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
	"strings"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ModuleCommand deal "command" module.
func ModuleCommand(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}
	// get connector
	conn, err := getConnector(ctx, options.Host, options.Variable)
	if err != nil {
		return "", fmt.Sprintf("failed to connector of %q error: %v", options.Host, err)
	}
	defer conn.Close(ctx)
	// command string
	command, err := variable.Extension2String(ha, options.Args)
	if err != nil {
		return "", err.Error()
	}
	// execute command
	var stdout, stderr string
	data, err := conn.ExecuteCommand(ctx, command)
	if err != nil {
		stderr = err.Error()
	}
	if data != nil {
		stdout = strings.TrimSpace(string(data))
	}

	return stdout, stderr
}
