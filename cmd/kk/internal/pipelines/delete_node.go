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
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/config"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/confirm"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/precheck"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/kubernetes"
	"github.com/kubesphere/kubekey/util/workflow/module"
	"github.com/kubesphere/kubekey/util/workflow/pipeline"
)

func DeleteNodePipeline(runtime *common.KubeRuntime) error {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&confirm.DeleteNodeConfirmModule{},
		&config.ModifyConfigModule{},
		&kubernetes.CompareConfigAndClusterInfoModule{},
		&kubernetes.DeleteKubeNodeModule{},
	}

	p := pipeline.Pipeline{
		Name:    "DeleteNodePipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func DeleteNode(args common.Argument) error {
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

	if err := DeleteNodePipeline(runtime); err != nil {
		return err
	}
	return nil
}
