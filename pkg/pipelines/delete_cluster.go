/*
 Copyright 2021 The KubeSphere Authors.

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

package pipelines

import (
	"github.com/kubesphere/kubekey/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/certs"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/k3s"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

func NewDeleteClusterPipeline(runtime *common.KubeRuntime) error {
	m := []module.Module{
		&confirm.DeleteClusterConfirmModule{},
		&kubernetes.ResetClusterModule{},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
	}

	p := pipeline.Pipeline{
		Name:    "DeleteClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func NewK3sDeleteClusterPipeline(runtime *common.KubeRuntime) error {
	m := []module.Module{
		&confirm.DeleteClusterConfirmModule{},
		&k3s.DeleteClusterModule{},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
	}

	p := pipeline.Pipeline{
		Name:    "K3sDeleteClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func DeleteCluster(args common.Argument) error {
	var loaderType string
	if args.FilePath != "" {
		loaderType = common.File
	} else {
		loaderType = common.AllInOne
	}

	runtime, err := common.NewKubeRuntime(loaderType, args)
	if err != nil {
		return err
	}

	switch runtime.Cluster.Kubernetes.Type {
	case common.K3s:
		if err := NewK3sDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	case common.Kubernetes:
		if err := NewDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	default:
		if err := NewDeleteClusterPipeline(runtime); err != nil {
			return err
		}
	}
	return nil
}
