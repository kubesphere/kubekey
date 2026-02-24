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

package internal

import (
	"context"
	"io"
	"io/fs"

	"k8s.io/klog/v2"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestConnector creates a mock implementation of the connector.Connector interface
// for testing purposes. The returned connector uses the provided stdout, stderr, and err
// to simulate the result of operations. All methods will behave as follows:
// - ExecuteCommand returns the preset stdout, stderr, and err values.
// - PutFile, FetchFile, Init, and Close all simply return the err value.
// This allows unit tests to simulate various outputs and errors from a module's connector-dependent calls.
func NewTestConnector(stdout, stderr string, err error) connector.Connector {
	return &testConnector{
		stdout: stdout, // Simulated standard output for ExecuteCommand
		stderr: stderr, // Simulated standard error output for ExecuteCommand
		err:    err,    // Simulated error to return from all methods
	}
}

// testConnector is a simple mock implementation of connector.Connector for use in tests.
// All network/file operations are stubbed out and return preset results.
type testConnector struct {
	stdout string // Output to simulate when running ExecuteCommand
	stderr string // Error output to simulate when running ExecuteCommand
	err    error  // Error to return from all connector methods
}

// Init simulates connector initialization. It always returns the preset error value.
func (t testConnector) Init(context.Context) error {
	return t.err
}

// Close simulates closing the connector. It always returns the preset error value.
func (t testConnector) Close(context.Context) error {
	return t.err
}

// PutFile simulates uploading a file to a remote machine. Always returns the preset error value.
func (t testConnector) PutFile(context.Context, []byte, string, fs.FileMode) error {
	return t.err
}

// FetchFile simulates fetching a file from a remote machine. Always returns the preset error value.
func (t testConnector) FetchFile(context.Context, string, io.Writer) error {
	return t.err
}

// ExecuteCommand simulates running a shell command on a remote host. It returns
// the preset stdout, stderr, and err values to allow precise test control over module results.
func (t testConnector) ExecuteCommand(context.Context, string) ([]byte, []byte, error) {
	return []byte(t.stdout), []byte(t.stderr), t.err
}

// NewTestVariable creates a new variable.Variable for testing purposes.
// It initializes a test playbook and client via _const.NewTestPlaybook, then
// creates an in-memory variable source. It merges the provided vars as remote variables
// for the specified hosts. This allows you to customize the per-host variable context
// for unit tests needing module execution context.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		// If creating the test playbook failed, log the error and continue (returned Variable may be nil)
		klog.ErrorS(err, "failed to create test playbook")
	}
	// Create a new variable in memory using the test client and playbook
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		// If creating the variable failed, log and return what we have (likely nil)
		klog.ErrorS(err, "failed to create variable")
	}
	// Set default values by merging the provided vars as remote variables for the hosts.
	if err := v.Merge(variable.MergeRemoteVariable(vars, hosts...)); err != nil {
		// If merging variables failed, log the error.
		klog.ErrorS(err, "failed to merge variable")
	}

	return v
}
