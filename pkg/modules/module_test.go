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
	"errors"
	"io"
	"io/fs"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type testVariable struct {
	value map[string]any
	err   error
}

func (v testVariable) Key() string {
	return "testModule"
}

func (v testVariable) Get(variable.GetFunc) (any, error) {
	return v.value, v.err
}

func (v *testVariable) Merge(variable.MergeFunc) error {
	v.value = map[string]any{
		"k": "v",
	}

	return nil
}

var successConnector = &testConnector{output: []byte("success")}
var failedConnector = &testConnector{
	copyErr:    errors.New("failed"),
	fetchErr:   errors.New("failed"),
	commandErr: errors.New("failed"),
}

var _ connector.Connector = &testConnector{}

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

func (t testConnector) Init(context.Context) error {
	return t.initErr
}

func (t testConnector) Close(context.Context) error {
	return t.closeErr
}

func (t testConnector) PutFile(context.Context, []byte, string, fs.FileMode) error {
	return t.copyErr
}

func (t testConnector) FetchFile(context.Context, string, io.Writer) error {
	return t.fetchErr
}

func (t testConnector) ExecuteCommand(context.Context, string) ([]byte, error) {
	return t.output, t.commandErr
}
