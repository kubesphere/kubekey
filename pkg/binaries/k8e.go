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

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/pkg/errors"
)

// K8eFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K8eFilesDownloadHTTP(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache) error {

	etcd := files.NewKubeBinary("etcd", arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path, kubeConf.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, kubekeyapiv1alpha2.DefaultCniVersion, path, kubeConf.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, kubekeyapiv1alpha2.DefaultHelmVersion, path, kubeConf.Arg.DownloadCommand)
	k8e := files.NewKubeBinary("k8e", arch, version, path, kubeConf.Arg.DownloadCommand)

	binaries := []*files.KubeBinary{k8e, helm, kubecni, etcd}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			p := binary.Path()
			if err := binary.SHA256Check(); err != nil {
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

func K8eArtifactBinariesDownload(manifest *common.ArtifactManifest, path, arch, version string) error {
	m := manifest.Spec

	etcd := files.NewKubeBinary("etcd", arch, m.Components.ETCD.Version, path, manifest.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, m.Components.CNI.Version, path, manifest.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, m.Components.Helm.Version, path, manifest.Arg.DownloadCommand)
	k8e := files.NewKubeBinary("k8e", arch, version, path, manifest.Arg.DownloadCommand)
	crictl := files.NewKubeBinary("crictl", arch, m.Components.Crictl.Version, path, manifest.Arg.DownloadCommand)
	binaries := []*files.KubeBinary{k8e, helm, kubecni, etcd}

	dockerArr := make([]*files.KubeBinary, 0, 0)
	dockerVersionMap := make(map[string]struct{})
	for _, c := range m.Components.ContainerRuntimes {
		var dockerVersion string
		if c.Type == common.Docker {
			dockerVersion = c.Version
		} else {
			dockerVersion = kubekeyapiv1alpha2.DefaultDockerVersion
		}
		if _, ok := dockerVersionMap[dockerVersion]; !ok {
			dockerVersionMap[dockerVersion] = struct{}{}
			docker := files.NewKubeBinary("docker", arch, dockerVersion, path, manifest.Arg.DownloadCommand)
			dockerArr = append(dockerArr, docker)
		}
	}

	binaries = append(binaries, dockerArr...)
	if m.Components.Crictl.Version != "" {
		binaries = append(binaries, crictl)
	}

	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path())).Run()
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
