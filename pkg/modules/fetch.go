/*
Copyright 2024 The KubeSphere Authors.

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
	"os"
	"path/filepath"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// ModuleFetch deal fetch module
func ModuleFetch(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}
	// check args
	args := variable.Extension2Variables(options.Args)
	srcParam, err := variable.StringVar(ha, args, "src")
	if err != nil {
		return "", "\"src\" in args should be string"
	}
	destParam, err := variable.StringVar(ha, args, "dest")
	if err != nil {
		return "", "\"dest\" in args should be string"
	}

	// get connector
	conn, err := getConnector(ctx, options.Host, options.Variable)
	if err != nil {
		return "", fmt.Sprintf("get connector error: %v", err)
	}
	defer conn.Close(ctx)

	// fetch file
	if _, err := os.Stat(filepath.Dir(destParam)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(destParam), os.ModePerm); err != nil {
			return "", fmt.Sprintf("failed to create dest dir: %v", err)
		}
	}

	destFile, err := os.Create(destParam)
	if err != nil {
		return "", err.Error()
	}
	defer destFile.Close()

	if err := conn.FetchFile(ctx, srcParam, destFile); err != nil {
		return "", fmt.Sprintf("failed to fetch file: %v", err)
	}

	return StdoutSuccess, ""
}
