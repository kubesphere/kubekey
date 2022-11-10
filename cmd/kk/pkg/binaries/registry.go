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

package binaries

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/cache"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/files"
)

// RegistryPackageDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func RegistryPackageDownloadHTTP(kubeConf *common.KubeConf, path, arch string, pipelineCache *cache.Cache) error {
	var binaries []*files.KubeBinary

	switch kubeConf.Cluster.Registry.Type {
	case common.Harbor:
		// TODO: Harbor only supports amd64, so there is no need to consider other architectures at present.
		harbor := files.NewKubeBinary("harbor", arch, kubekeyapiv1alpha2.DefaultHarborVersion, path, kubeConf.Arg.DownloadCommand)
		compose := files.NewKubeBinary("compose", arch, kubekeyapiv1alpha2.DefaultDockerComposeVersion, path, kubeConf.Arg.DownloadCommand)
		docker := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path, kubeConf.Arg.DownloadCommand)
		binaries = []*files.KubeBinary{harbor, docker, compose}
	default:
		registry := files.NewKubeBinary("registry", arch, kubekeyapiv1alpha2.DefaultRegistryVersion, path, kubeConf.Arg.DownloadCommand)
		binaries = []*files.KubeBinary{registry}
	}

	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s  ...", arch, binary.ID, binary.Version)
		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.GetCmd(), err)
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}

func RegistryBinariesDownload(manifest *common.ArtifactManifest, path, arch string) error {

	m := manifest.Spec
	binaries := make([]*files.KubeBinary, 0, 0)

	if m.Components.DockerRegistry.Version != "" {
		registry := files.NewKubeBinary("registry", arch, kubekeyapiv1alpha2.DefaultRegistryVersion, path, manifest.Arg.DownloadCommand)
		binaries = append(binaries, registry)
	}

	if m.Components.Harbor.Version != "" {
		harbor := files.NewKubeBinary("harbor", arch, kubekeyapiv1alpha2.DefaultHarborVersion, path, manifest.Arg.DownloadCommand)
		if arch == "amd64" {
			binaries = append(binaries, harbor)
		} else {
			logger.Log.Warningf("Harbor only supports amd64, the KubeKey artifact will not contain the Harbor %s", arch)
		}
	}

	if m.Components.DockerCompose.Version != "" {
		compose := files.NewKubeBinary("compose", arch, kubekeyapiv1alpha2.DefaultDockerComposeVersion, path, manifest.Arg.DownloadCommand)
		containerManager := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path, manifest.Arg.DownloadCommand)
		if arch == "amd64" {
			binaries = append(binaries, compose)
			binaries = append(binaries, containerManager)
		}
	}

	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.GetCmd(), err)
		}
	}
	return nil
}
