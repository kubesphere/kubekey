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

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

const (
	// StdoutSuccess is the standard message indicating a successful module execution.
	StdoutSuccess = "success"
	// StdoutFailed is the standard message indicating a failed module execution.
	StdoutFailed = "failed"
	// StdoutSkip is the standard message indicating a skipped module execution.
	StdoutSkip = "skip"
	// StdoutTrue is the standard message indicating a boolean true result (used in bool/assert modules).
	StdoutTrue = "True"
	// StdoutFalse is the standard message indicating a boolean false result (used in bool/assert modules).
	StdoutFalse = "False"
)

// ModuleExecFunc defines the function signature for executing a module.
// It takes a context and ExecOptions, and returns stdout and stderr strings.
type ModuleExecFunc func(ctx context.Context, options ExecOptions) (stdout string, stderr string)

// ExecOptions represents options for module execution.
type ExecOptions struct {
	// Args contains the defined arguments for the module.
	Args runtime.RawExtension
	// Host specifies which host to execute the module on.
	Host string
	// Variable provides the variables needed by the module.
	variable.Variable
	// Task is the task to be executed.
	Task kkcorev1alpha1.Task
	// Playbook is the playbook to be executed.
	Playbook kkcorev1.Playbook
	// LogOutput is the output writer for module logs.
	LogOutput io.Writer
}

// getAllVariables retrieves all variables for the specified host in ExecOptions.
// Returns a map of variables or an error if retrieval fails.
func (o ExecOptions) getAllVariables() (map[string]any, error) {
	ha, err := o.Variable.Get(variable.GetAllVariable(o.Host))
	if err != nil {
		return nil, err
	}

	vd, ok := ha.(map[string]any)
	if !ok {
		return nil, errors.Errorf("host: %s variable is not a map", o.Host)
	}

	return vd, nil
}

// getConnector returns a connector for the specified host in ExecOptions.
// If the task has a DelegateTo field, it parses and uses that host instead.
// It first checks if a connector is already present in the context, otherwise creates a new one.
func (o ExecOptions) getConnector(ctx context.Context) (connector.Connector, error) {
	var conn connector.Connector
	var err error

	ha, err := o.getAllVariables()
	if err != nil {
		return nil, err
	}

	host := o.Host
	if o.Task.Spec.DelegateTo != "" {
		host, err = tmpl.ParseFunc(ha, o.Task.Spec.DelegateTo, func(b []byte) string { return string(b) })
		if err != nil {
			return nil, errors.Errorf("failed to delegate %q to %q", o.Host, o.Task.Spec.DelegateTo)
		}
	}
	if val := ctx.Value(ConnKey); val != nil {
		if vd, ok := val.(connector.Connector); ok {
			conn = vd
		}
	} else {
		conn, err = connector.NewConnector(host, o.Variable)
		if err != nil {
			return conn, err
		}
	}

	if err = conn.Init(ctx); err != nil {
		return conn, err
	}

	return conn, nil
}

// module is a registry mapping module names to their execution functions.
var module = make(map[string]ModuleExecFunc)

// RegisterModule registers a module execution function under the given module name.
// Returns an error if the module name is already registered.
func RegisterModule(moduleName string, exec ModuleExecFunc) error {
	if _, ok := module[moduleName]; ok {
		return errors.Errorf("module %s is exist", moduleName)
	}

	module[moduleName] = exec

	return nil
}

// FindModule retrieves a registered module execution function by its name.
// Returns nil if the module is not found.
func FindModule(moduleName string) ModuleExecFunc {
	return module[moduleName]
}

// key is an unexported type used for context keys in this package.
type key struct{}

// ConnKey is the context key for storing/retrieving a connector in context.Context.
var ConnKey = &key{}
