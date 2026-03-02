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

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

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
	// +optional
	LogOutput io.Writer
}

// GetAllVariables retrieves all variables for the specified host in Execinternal.
func (o ExecOptions) GetAllVariables() (map[string]any, error) {
	ha, err := o.Get(variable.GetAllVariable(o.Host))
	if err != nil {
		return nil, err
	}

	vd, ok := ha.(map[string]any)
	if !ok {
		return nil, errors.Errorf("host: %s variable is not a map", o.Host)
	}

	return vd, nil
}

// GetConnector returns a connector for the specified host in Execinternal.
func (o ExecOptions) GetConnector(ctx context.Context) (connector.Connector, error) {
	var conn connector.Connector
	var err error

	ha, err := o.GetAllVariables()
	if err != nil {
		return nil, err
	}

	host := o.Host
	if o.Task.Spec.DelegateTo != "" {
		host, err = tmpl.ParseFunc(ha, o.Task.Spec.DelegateTo, tmpl.StringFunc)
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

// Error is a simple error type for module registration errors.
type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}

// key is an unexported type used for context keys in this package.
type key struct{}

// ConnKey is the context key for storing/retrieving a connector in context.Context.
var ConnKey = &key{}

// ModuleExecFunc defines the function signature for executing a module.
type ModuleExecFunc func(ctx context.Context, opts ExecOptions) (stdout string, stderr string, err error)

// module is a registry mapping module names to their execution functions.
var moduleRegistry = make(map[string]ModuleExecFunc)

// RegisterModule registers a module execution function under the given module name.
func RegisterModule(exec ModuleExecFunc, moduleName ...string) error {
	for _, name := range moduleName {
		if _, ok := moduleRegistry[name]; ok {
			return &Error{Msg: "module " + name + " already exists"}
		}
		moduleRegistry[name] = exec
	}
	return nil
}

// FindModule retrieves a registered module execution function by its name.
func FindModule(moduleName string) ModuleExecFunc {
	return moduleRegistry[moduleName]
}
