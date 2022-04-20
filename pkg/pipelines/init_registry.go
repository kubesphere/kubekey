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
	"github.com/kubesphere/kubekey/pkg/artifact"
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os"
	"github.com/kubesphere/kubekey/pkg/bootstrap/precheck"
	"github.com/kubesphere/kubekey/pkg/bootstrap/registry"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
	"github.com/kubesphere/kubekey/pkg/filesystem"
)

func NewInitRegistryPipeline(runtime *common.KubeRuntime) error {
	noArtifact := runtime.Arg.Artifact == ""

	m := []module.Module{
		&precheck.GreetingsModule{},
		&artifact.UnArchiveModule{Skip: noArtifact},
		&binaries.RegistryPackageModule{},
		&os.ConfigureOSModule{},
		&registry.RegistryCertsModule{},
		&registry.InstallRegistryModule{},
		&filesystem.ChownWorkDirModule{},
	}

	p := pipeline.Pipeline{
		Name:    "InitRegistryPipeline",
		Modules: m,
		Runtime: runtime,
	}
	if err := p.Start(); err != nil {
		return err
	}
	return nil
}

func InitRegistry(args common.Argument, downloadCmd string) error {
	args.DownloadCommand = func(path, url string) string {
		// this is an extension point for downloading tools, for example users can set the timeout, proxy or retry under
		// some poor network environment. Or users even can choose another cli, it might be wget.
		// perhaps we should have a build-in download function instead of totally rely on the external one
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

	if err := NewInitRegistryPipeline(runtime); err != nil {
		return err
	}
	return nil
}
