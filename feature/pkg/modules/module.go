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

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
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

// ExecOptions for module
type ExecOptions struct {
	// the defined Args for module
	Args runtime.RawExtension
	// which Host to execute
	Host string
	// the variable module need
	variable.Variable
	// the task to be executed
	Task kkcorev1alpha1.Task
	// the pipeline to be executed
	Pipeline kkcorev1.Pipeline
}

func (o ExecOptions) getAllVariables() (map[string]any, error) {
	ha, err := o.Variable.Get(variable.GetAllVariable(o.Host))
	if err != nil {
		return nil, fmt.Errorf("failed to get host %s variable: %w", o.Host, err)
	}

	vd, ok := ha.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("host: %s variable is not a map", o.Host)
	}

	return vd, nil
}

var module = make(map[string]ModuleExecFunc)

// RegisterModule register module
func RegisterModule(moduleName string, exec ModuleExecFunc) error {
	if _, ok := module[moduleName]; ok {
		return fmt.Errorf("module %s is exist", moduleName)
	}

	module[moduleName] = exec

	return nil
}

// FindModule by module name which has register.
func FindModule(moduleName string) ModuleExecFunc {
	return module[moduleName]
}

func init() {
	utilruntime.Must(RegisterModule("assert", ModuleAssert))
	utilruntime.Must(RegisterModule("command", ModuleCommand))
	utilruntime.Must(RegisterModule("shell", ModuleCommand))
	utilruntime.Must(RegisterModule("copy", ModuleCopy))
	utilruntime.Must(RegisterModule("fetch", ModuleFetch))
	utilruntime.Must(RegisterModule("debug", ModuleDebug))
	utilruntime.Must(RegisterModule("template", ModuleTemplate))
	utilruntime.Must(RegisterModule("set_fact", ModuleSetFact))
	utilruntime.Must(RegisterModule("gen_cert", ModuleGenCert))
	utilruntime.Must(RegisterModule("image", ModuleImage))
}

// ConnKey for connector which store in context
var ConnKey = struct{}{}

func getConnector(ctx context.Context, host string, data map[string]any) (connector.Connector, error) {
	var conn connector.Connector
	var err error

	if v := ctx.Value(ConnKey); v != nil {
		if vd, ok := v.(connector.Connector); ok {
			conn = vd
		}
	} else {
		connectorVars := make(map[string]any)

		if c1, ok := data[_const.VariableConnector]; ok {
			if c2, ok := c1.(map[string]any); ok {
				connectorVars = c2
			}
		}

		conn, err = connector.NewConnector(host, connectorVars)
		if err != nil {
			return conn, err
		}
	}

	if err = conn.Init(ctx); err != nil {
		klog.V(4).ErrorS(err, "failed to init connector")

		return conn, err
	}

	return conn, nil
}
