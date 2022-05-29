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
	"fmt"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/cmd/kk/internal/artifact"
	"github.com/kubesphere/kubekey/cmd/kk/internal/binaries"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/confirm"
	"github.com/kubesphere/kubekey/cmd/kk/internal/bootstrap/precheck"
	"github.com/kubesphere/kubekey/cmd/kk/internal/common"
	"github.com/kubesphere/kubekey/cmd/kk/internal/filesystem"
	"github.com/kubesphere/kubekey/cmd/kk/internal/images"
	"github.com/kubesphere/kubekey/util/workflow/module"
	"github.com/kubesphere/kubekey/util/workflow/pipeline"
)

func NewArtifactExportPipeline(runtime *common.ArtifactRuntime) error {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&confirm.CheckFileExistModule{FileName: runtime.Arg.Output},
		&images.CopyImagesToLocalModule{},
		&binaries.ArtifactBinariesModule{},
		&artifact.RepositoryModule{},
		&artifact.ArchiveModule{},
		&filesystem.ChownOutputModule{},
		&filesystem.ChownWorkDirModule{},
	}

	p := pipeline.Pipeline{
		Name:            "ArtifactExportPipeline",
		Modules:         m,
		Runtime:         runtime,
		ModulePostHooks: nil,
	}
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}

func NewK3sArtifactExportPipeline(runtime *common.ArtifactRuntime) error {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&confirm.CheckFileExistModule{FileName: runtime.Arg.Output},
		&images.CopyImagesToLocalModule{},
		&binaries.K3sArtifactBinariesModule{},
		&artifact.RepositoryModule{},
		&artifact.ArchiveModule{},
		&filesystem.ChownOutputModule{},
		&filesystem.ChownWorkDirModule{},
	}

	p := pipeline.Pipeline{
		Name:            "K3sArtifactExportPipeline",
		Modules:         m,
		Runtime:         runtime,
		ModulePostHooks: nil,
	}
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}

func ArtifactExport(args common.ArtifactArgument, downloadCmd string) error {
	args.DownloadCommand = func(path, url string) string {
		// this is an extension point for downloading tools, for example users can set the timeout, proxy or retry under
		// some poor network environment. Or users even can choose another cli, it might be wget.
		// perhaps we should have a build-in download function instead of totally rely on the external one
		return fmt.Sprintf(downloadCmd, path, url)
	}

	runtime, err := common.NewArtifactRuntime(args)
	if err != nil {
		return err
	}

	if len(runtime.Spec.KubernetesDistributions) == 0 {
		return errors.New("the length of kubernetes distributions can't be 0")
	}

	pre := runtime.Spec.KubernetesDistributions[0].Type
	for _, t := range runtime.Spec.KubernetesDistributions {
		if t.Type != pre {
			return errors.New("all the types of kubernetes distributions can't be different")
		}
	}

	switch runtime.Spec.KubernetesDistributions[0].Type {
	case common.K3s:
		if err := NewK3sArtifactExportPipeline(runtime); err != nil {
			return err
		}
	case common.Kubernetes:
		if err := NewArtifactExportPipeline(runtime); err != nil {
			return err
		}
	default:
		if err := NewArtifactExportPipeline(runtime); err != nil {
			return err
		}
	}

	return nil
}
