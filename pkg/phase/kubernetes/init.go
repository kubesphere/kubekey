/*
 Copyright 2022 The KubeSphere Authors.

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

package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v2/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/v2/pkg/common"
	"github.com/kubesphere/kubekey/v2/pkg/core/module"
	"github.com/kubesphere/kubekey/v2/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/v2/pkg/kubernetes"
	"github.com/kubesphere/kubekey/v2/pkg/phase/confirm"
	"github.com/kubesphere/kubekey/v2/pkg/plugins/dns"
)

func NewCreateInitClusterPipeline(runtime *common.KubeRuntime) error {
	m := []module.Module{
		&precheck.NodePreCheckModule{},
		&kubernetes.StatusModule{},
		&confirm.CreateK8sConfirmModule{},
		&InstallKubeletModule{},
		&kubernetes.InitKubernetesModule{},
		&dns.ClusterDNSModule{},
	}

	p := pipeline.Pipeline{
		Name:    "CreateInitClusterPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func CreateInitCluster(args common.Argument) error {

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
	case common.Kubernetes:
		if err := NewCreateInitClusterPipeline(runtime); err != nil {
			return err
		}
	default:
		return errors.New("unsupported cluster kubernetes type")
	}

	return nil
}
