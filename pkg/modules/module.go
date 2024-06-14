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
	"k8s.io/klog/v2"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/v4/pkg/connector"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
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
	kubekeyv1alpha1.Task
	// the pipeline to be executed
	kubekeyv1.Pipeline
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
	RegisterModule("assert", ModuleAssert)
	RegisterModule("command", ModuleCommand)
	RegisterModule("shell", ModuleCommand)
	RegisterModule("copy", ModuleCopy)
	RegisterModule("fetch", ModuleFetch)
	RegisterModule("debug", ModuleDebug)
	RegisterModule("template", ModuleTemplate)
	RegisterModule("set_fact", ModuleSetFact)
	RegisterModule("gen_cert", ModuleGenCert)
	RegisterModule("image", ModuleImage)
}

// ConnKey for connector which store in context
var ConnKey = struct{}{}

func getConnector(ctx context.Context, host string, data map[string]any) (connector.Connector, error) {
	var conn connector.Connector
	var err error
	if v := ctx.Value(ConnKey); v != nil {
		conn = v.(connector.Connector)
	} else {
		conn, err = connector.NewConnector(host, data)
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
