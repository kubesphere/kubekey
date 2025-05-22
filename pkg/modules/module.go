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
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// message for stdout
const (
	// StdoutSuccess message for common module
	StdoutSuccess = "success"
	StdoutSkip    = "skip"

	// StdoutTrue for bool module
	StdoutTrue = "True"
	// StdoutFalse for bool module
	StdoutFalse = "False"
)

// ModuleExecFunc exec module
type ModuleExecFunc func(ctx context.Context, options ExecOptions) (stdout string, stderr string)

// ExecOptions represents options for module execution
type ExecOptions struct {
	// the defined Args for module
	Args runtime.RawExtension
	// which Host to execute
	Host string
	// the variable module need
	variable.Variable
	// the task to be executed
	Task kkcorev1alpha1.Task
	// the playbook to be executed
	Playbook kkcorev1.Playbook
	// the logout for module
	LogOutput io.Writer
}

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

var module = make(map[string]ModuleExecFunc)

// RegisterModule register module
func RegisterModule(moduleName string, exec ModuleExecFunc) error {
	if _, ok := module[moduleName]; ok {
		return errors.Errorf("module %s is exist", moduleName)
	}

	module[moduleName] = exec

	return nil
}

// FindModule by module name which has register.
func FindModule(moduleName string) ModuleExecFunc {
	return module[moduleName]
}

type key struct{}

// ConnKey for connector which store in context
var ConnKey = &key{}

func getConnector(ctx context.Context, host string, v variable.Variable) (connector.Connector, error) {
	var conn connector.Connector
	var err error

	if val := ctx.Value(ConnKey); val != nil {
		if vd, ok := val.(connector.Connector); ok {
			conn = vd
		}
	} else {
		conn, err = connector.NewConnector(host, v)
		if err != nil {
			return conn, err
		}
	}

	if err = conn.Init(ctx); err != nil {
		return conn, err
	}

	return conn, nil
}
