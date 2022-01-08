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
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/pkg/errors"
	"io"
	"os"
	"os/exec"
)

// RegistryPackageDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func RegistryPackageDownloadHTTP(kubeConf *common.KubeConf, filepath, arch string, pipelineCache *cache.Cache) error {
	kkzone := os.Getenv("KKZONE")

	binaries := []files.KubeBinary{}

	switch kubeConf.Cluster.Registry.Type {
	case common.Harbor:
		harbor := files.KubeBinary{Name: "harbor", Arch: arch, Version: kubekeyapiv1alpha2.DefaultHarborVersion}
		harbor.Path = fmt.Sprintf("%s/harbor-offline-installer-%s.tgz", filepath, kubekeyapiv1alpha2.DefaultHarborVersion)
		// TODO: Harbor only supports amd64, so there is no need to consider other architectures at present.
		docker := files.KubeBinary{Name: "docker", Arch: arch, Version: kubekeyapiv1alpha2.DefaultDockerVersion}
		docker.Path = fmt.Sprintf("%s/docker-%s.tgz", filepath, kubekeyapiv1alpha2.DefaultDockerVersion)
		compose := files.KubeBinary{Name: "compose", Arch: arch, Version: kubekeyapiv1alpha2.DefaultDockerComposeVersion}
		compose.Path = fmt.Sprintf("%s/docker-compose-linux-x86_64", filepath)

		if kkzone == "cn" {
			harbor.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", harbor.Version, harbor.Version)
			docker.Url = fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/x86_64/docker-%s.tgz", docker.Version)
			compose.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64", compose.Version)
		} else {
			harbor.Url = fmt.Sprintf("https://github.com/goharbor/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", harbor.Version, harbor.Version)
			docker.Url = fmt.Sprintf("https://download.docker.com/linux/static/stable/x86_64/docker-%s.tgz", docker.Version)
			compose.Url = fmt.Sprintf("https://github.com/docker/compose/releases/download/%s/docker-compose-linux-x86_64", compose.Version)
		}

		harbor.GetCmd = kubeConf.Arg.DownloadCommand(harbor.Path, harbor.Url)
		docker.GetCmd = kubeConf.Arg.DownloadCommand(docker.Path, docker.Url)
		compose.GetCmd = kubeConf.Arg.DownloadCommand(compose.Path, compose.Url)

		binaries = []files.KubeBinary{harbor, docker, compose}
	default:
		registry := files.KubeBinary{Name: "registry", Arch: arch, Version: kubekeyapiv1alpha2.DefaultRegistryVersion}
		registry.Path = fmt.Sprintf("%s/registry-%s-linux-%s.tar.gz", filepath, kubekeyapiv1alpha2.DefaultRegistryVersion, arch)
		if kkzone == "cn" {
			registry.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/registry/%s/registry-%s-linux-%s.tar.gz", kubekeyapiv1alpha2.DefaultRegistryVersion, kubekeyapiv1alpha2.DefaultRegistryVersion, registry.Arch)
		} else {
			registry.Url = fmt.Sprintf("https://github.com/kubesphere/kubekey/releases/download/v2.0.0-alpha.1/registry-%s-linux-%s.tar.gz", kubekeyapiv1alpha2.DefaultRegistryVersion, registry.Arch)
		}
		registry.GetCmd = kubeConf.Arg.DownloadCommand(registry.Path, registry.Url)
		binaries = []files.KubeBinary{registry}
	}

	binariesMap := make(map[string]files.KubeBinary)
	for _, binary := range binaries {
		logger.Log.Messagef(common.LocalHost, "downloading %s ...", binary.Name)
		binariesMap[binary.Name] = binary
		if util.IsExist(binary.Path) {
			// download it again if it's incorrect
			if err := SHA256Check(binary); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
			} else {
				continue
			}
		}

		for i := 5; i > 0; i-- {
			cmd := exec.Command("/bin/sh", "-c", binary.GetCmd)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.Name, binary.GetCmd, err)
			}
			cmd.Stderr = cmd.Stdout

			if err = cmd.Start(); err != nil {
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.Name, binary.GetCmd, err)
			}
			for {
				tmp := make([]byte, 1024)
				_, err := stdout.Read(tmp)
				fmt.Print(string(tmp)) // Get the output from the pipeline in real time and print it to the terminal
				if errors.Is(err, io.EOF) {
					break
				} else if err != nil {
					logger.Log.Errorln(err)
					break
				}
			}
			if err = cmd.Wait(); err != nil {
				if kkzone != "cn" {
					logger.Log.Warningln("Having a problem with accessing https://storage.googleapis.com? You can try again after setting environment 'export KKZONE=cn'")
				}
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.Name, binary.GetCmd, err)
			}

			if err := SHA256Check(binary); err != nil {
				if i == 1 {
					return err
				}
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
				continue
			}
			break
		}
	}

	pipelineCache.Set(common.KubeBinaries, binariesMap)
	return nil
}
