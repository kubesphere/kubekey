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
	"io"
	"io/fs"

	"github.com/cockroachdb/errors"
	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

var successConnector = &testConnector{stdout: []byte("success")}
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
	stdout     []byte
	stderr     []byte
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

func (t testConnector) ExecuteCommand(context.Context, string) ([]byte, []byte, error) {
	return t.stdout, t.stderr, t.commandErr
}

// newTestVariable creates a new variable.Variable for testing purposes.
// It initializes a test playbook and client, creates a new in-memory variable source,
// and merges the provided vars as remote variables for the specified hosts.
func newTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		klog.Error(err)
	}
	// Create a new variable in memory using the test client and playbook
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		klog.Error(err)
	}
	// Set default values by merging the provided vars as remote variables for the hosts
	if err := v.Merge(variable.MergeRemoteVariable(vars, hosts...)); err != nil {
		klog.Error(err)
	}

	return v
}
