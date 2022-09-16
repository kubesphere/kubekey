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

package alpha

import (
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/pkg/alpha/confirm"
	"github.com/kubesphere/kubekey/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
)

func NewCreateKubeSpherePipeline(runtime *common.KubeRuntime) error {

	m := []module.Module{
		&precheck.NodePreCheckModule{},
		&confirm.CreateKsConfirmModule{},
		&kubesphere.DeployModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
		&kubesphere.CheckResultModule{Skip: !runtime.Cluster.KubeSphere.Enabled},
	}

	p := pipeline.Pipeline{
		Name:    "CreateKubeSpherePipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func CreateKubeSphere(args common.Argument) error {

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
		if err := NewCreateKubeSpherePipeline(runtime); err != nil {
			return err
		}
	default:
		return errors.New("unsupported cluster kubernetes type")
	}

	return nil
}
