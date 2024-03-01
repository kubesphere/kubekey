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

package binaries

import (
	"fmt"
	"os/exec"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/cache"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
	"github.com/pkg/errors"
)

// K8sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K8sFilesDownloadHTTP(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache) error {

	etcd := files.NewKubeBinary("etcd", arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path, kubeConf.Arg.DownloadCommand)
	kubeadm := files.NewKubeBinary("kubeadm", arch, version, path, kubeConf.Arg.DownloadCommand)
	kubelet := files.NewKubeBinary("kubelet", arch, version, path, kubeConf.Arg.DownloadCommand)
	kubectl := files.NewKubeBinary("kubectl", arch, version, path, kubeConf.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, kubekeyapiv1alpha2.DefaultCniVersion, path, kubeConf.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, kubekeyapiv1alpha2.DefaultHelmVersion, path, kubeConf.Arg.DownloadCommand)
	docker := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path, kubeConf.Arg.DownloadCommand)
	criDockerd := files.NewKubeBinary("cri-dockerd", arch, kubekeyapiv1alpha2.DefaultCriDockerdVersion, path, kubeConf.Arg.DownloadCommand)
	crictl := files.NewKubeBinary("crictl", arch, kubekeyapiv1alpha2.DefaultCrictlVersion, path, kubeConf.Arg.DownloadCommand)
	containerd := files.NewKubeBinary("containerd", arch, kubekeyapiv1alpha2.DefaultContainerdVersion, path, kubeConf.Arg.DownloadCommand)
	runc := files.NewKubeBinary("runc", arch, kubekeyapiv1alpha2.DefaultRuncVersion, path, kubeConf.Arg.DownloadCommand)
	calicoctl := files.NewKubeBinary("calicoctl", arch, kubekeyapiv1alpha2.DefaultCalicoVersion, path, kubeConf.Arg.DownloadCommand)

	binaries := []*files.KubeBinary{kubeadm, kubelet, kubectl, helm, kubecni, crictl, etcd}

	if kubeConf.Cluster.Kubernetes.ContainerManager == kubekeyapiv1alpha2.Docker {
		binaries = append(binaries, docker)
		if kubeConf.Cluster.Kubernetes.IsAtLeastV124() && kubeConf.Cluster.Kubernetes.ContainerManager == common.Docker {
			binaries = append(binaries, criDockerd)
		}
	} else if kubeConf.Cluster.Kubernetes.ContainerManager == kubekeyapiv1alpha2.Containerd {
		binaries = append(binaries, containerd, runc)
	}

	if kubeConf.Cluster.Network.Plugin == "calico" {
		binaries = append(binaries, calicoctl)
	}

	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Log.Messagef(common.LocalHost, "%s exists", binary.ID)
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.GetCmd(), err)
		}
	}

	if kubeConf.Cluster.KubeSphere.Version == "v2.1.1" {
		logger.Log.Infoln(fmt.Sprintf("Downloading %s ...", "helm2"))
		if util.IsExist(fmt.Sprintf("%s/helm2", helm.BaseDir)) == false {
			cmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm2", helm.BaseDir),
				fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, "v2.16.9"))
			if output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
				fmt.Println(string(output))
				return errors.Wrap(err, "Failed to download helm2 binary")
			}
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}

func KubernetesArtifactBinariesDownload(manifest *common.ArtifactManifest, path, arch, k8sVersion string) error {
	m := manifest.Spec

	etcd := files.NewKubeBinary("etcd", arch, m.Components.ETCD.Version, path, manifest.Arg.DownloadCommand)
	kubeadm := files.NewKubeBinary("kubeadm", arch, k8sVersion, path, manifest.Arg.DownloadCommand)
	kubelet := files.NewKubeBinary("kubelet", arch, k8sVersion, path, manifest.Arg.DownloadCommand)
	kubectl := files.NewKubeBinary("kubectl", arch, k8sVersion, path, manifest.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, m.Components.CNI.Version, path, manifest.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, m.Components.Helm.Version, path, manifest.Arg.DownloadCommand)
	crictl := files.NewKubeBinary("crictl", arch, m.Components.Crictl.Version, path, manifest.Arg.DownloadCommand)
	calicoctl := files.NewKubeBinary("calicoctl", arch, m.Components.Calicoctl.Version, path, manifest.Arg.DownloadCommand)
	binaries := []*files.KubeBinary{kubeadm, kubelet, kubectl, helm, kubecni, etcd, calicoctl}

	containerManagerArr := make([]*files.KubeBinary, 0, 0)
	containerManagerVersion := make(map[string]struct{})
	for _, c := range m.Components.ContainerRuntimes {
		if _, ok := containerManagerVersion[c.Type+c.Version]; !ok {
			containerManagerVersion[c.Type+c.Version] = struct{}{}
			containerManager := files.NewKubeBinary(c.Type, arch, c.Version, path, manifest.Arg.DownloadCommand)
			containerManagerArr = append(containerManagerArr, containerManager)
			if c.Type == "containerd" {
				runc := files.NewKubeBinary("runc", arch, kubekeyapiv1alpha2.DefaultRuncVersion, path, manifest.Arg.DownloadCommand)
				containerManagerArr = append(containerManagerArr, runc)
			}
		}
	}

	binaries = append(binaries, containerManagerArr...)
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

// CriDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func CriDownloadHTTP(kubeConf *common.KubeConf, path, arch string, pipelineCache *cache.Cache) error {

	binaries := []*files.KubeBinary{}
	switch kubeConf.Arg.Type {
	case common.Docker:
		docker := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path, kubeConf.Arg.DownloadCommand)
		binaries = append(binaries, docker)
	case common.Containerd:
		containerd := files.NewKubeBinary("containerd", arch, kubekeyapiv1alpha2.DefaultContainerdVersion, path, kubeConf.Arg.DownloadCommand)
		runc := files.NewKubeBinary("runc", arch, kubekeyapiv1alpha2.DefaultRuncVersion, path, kubeConf.Arg.DownloadCommand)
		crictl := files.NewKubeBinary("crictl", arch, kubekeyapiv1alpha2.DefaultCrictlVersion, path, kubeConf.Arg.DownloadCommand)
		binaries = append(binaries, containerd, runc, crictl)
	default:
	}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Log.Messagef(common.LocalHost, "downloading %s %s %s ...", arch, binary.ID, binary.Version)

		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Log.Messagef(common.LocalHost, "%s exists", binary.ID)
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
