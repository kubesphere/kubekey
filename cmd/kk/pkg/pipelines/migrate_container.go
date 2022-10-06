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

package pipelines

import (
	"fmt"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/binaries"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/bootstrap/confirm"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/container"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/module"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/pipeline"
)

func MigrateCriPipeline(runtime *common.KubeRuntime) error {
	fmt.Println("MigrateContainerdPipeline called")
	m := []module.Module{
		&precheck.GreetingsModule{},
		&confirm.MigrateCriConfirmModule{},
		&binaries.CriBinariesModule{},
		&container.CriMigrateModule{},
	}
	p := pipeline.Pipeline{
		Name:    "MigrateContainerdPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func MigrateCri(args common.Argument, downloadCmd string) error {
	args.DownloadCommand = func(path, url string) string {
		return fmt.Sprintf(downloadCmd, path, url)
	}
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
	if err := MigrateCriPipeline(runtime); err != nil {
		return err
	}
	return nil
}
