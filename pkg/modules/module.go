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

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

// message for stdout
const (
	// stdoutSuccess message for common module
	stdoutSuccess = "success"
	stdoutSkip    = "skip"

	// stdoutTrue for bool module
	stdoutTrue = "True"
	// stdoutFalse for bool module
	stdoutFalse = "False"
)

type ModuleExecFunc func(ctx context.Context, options ExecOptions) (stdout string, stderr string)

type ExecOptions struct {
	// the defined Args for module
	Args runtime.RawExtension
	// which Host to execute
	Host string
	// the variable module need
	variable.Variable
	// the task to be executed
	Task kubekeyv1alpha1.Task
	// the pipeline to be executed
	Pipeline kubekeyv1.Pipeline
}

var module = make(map[string]ModuleExecFunc)

func RegisterModule(moduleName string, exec ModuleExecFunc) error {
	if _, ok := module[moduleName]; ok {
		return fmt.Errorf("module %s is exist", moduleName)
	}
	module[moduleName] = exec
	return nil
}

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
		conn = v.(connector.Connector)
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
