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
	"github.com/kubesphere/kubekey/pkg/artifact"
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/pipeline"
)

func NewArtifactExportPipeline(runtime *common.ArtifactRuntime) error {
	m := []module.Module{
		&artifact.ImagesModule{},
		&binaries.ArtifactBinariesModule{},
		&artifact.RepositoryModule{},
		&artifact.ArchiveModule{},
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

	switch runtime.Spec.KubernetesDistribution.Type {
	case common.K3s:
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
