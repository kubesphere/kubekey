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
	"io"
	"io/fs"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type testVariable struct {
	value map[string]any
	err   error
}

func (v testVariable) Key() string {
	return "testModule"
}

func (v testVariable) Get(f variable.GetFunc) (any, error) {
	return v.value, v.err
}

func (v *testVariable) Merge(f variable.MergeFunc) error {
	v.value = map[string]any{
		"k": "v",
	}
	return nil
}

var successConnector = &testConnector{output: []byte("success")}
var failedConnector = &testConnector{
	copyErr:    fmt.Errorf("failed"),
	fetchErr:   fmt.Errorf("failed"),
	commandErr: fmt.Errorf("failed"),
}

type testConnector struct {
	// return for init
	initErr error
	// return for close
	closeErr error
	// return for copy
	copyErr error
	// return for fetch
	fetchErr error
	// return for command
	output     []byte
	commandErr error
}

func (t testConnector) Init(ctx context.Context) error {
	return t.initErr
}

func (t testConnector) Close(ctx context.Context) error {
	return t.closeErr
}

func (t testConnector) PutFile(ctx context.Context, local []byte, remoteFile string, mode fs.FileMode) error {
	return t.copyErr
}

func (t testConnector) FetchFile(ctx context.Context, remoteFile string, local io.Writer) error {
	return t.fetchErr
}

func (t testConnector) ExecuteCommand(ctx context.Context, cmd string) ([]byte, error) {
	return t.output, t.commandErr
}
