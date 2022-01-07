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
	"crypto/sha256"
	"fmt"
	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/cache"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// K8sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K8sFilesDownloadHTTP(kubeConf *common.KubeConf, filepath, version, arch string, pipelineCache *cache.Cache) error {
	kkzone := os.Getenv("KKZONE")
	etcd := files.KubeBinary{Name: "etcd", Arch: arch, Version: kubekeyapiv1alpha2.DefaultEtcdVersion}
	kubeadm := files.KubeBinary{Name: "kubeadm", Arch: arch, Version: version}
	kubelet := files.KubeBinary{Name: "kubelet", Arch: arch, Version: version}
	kubectl := files.KubeBinary{Name: "kubectl", Arch: arch, Version: version}
	kubecni := files.KubeBinary{Name: "kubecni", Arch: arch, Version: kubekeyapiv1alpha2.DefaultCniVersion}
	helm := files.KubeBinary{Name: "helm", Arch: arch, Version: kubekeyapiv1alpha2.DefaultHelmVersion}
	docker := files.KubeBinary{Name: "docker", Arch: arch, Version: kubekeyapiv1alpha2.DefaultDockerVersion}
	crictl := files.KubeBinary{Name: "crictl", Arch: arch, Version: kubekeyapiv1alpha2.DefaultCrictlVersion}

	etcd.Path = fmt.Sprintf("%s/etcd-%s-linux-%s.tar.gz", filepath, kubekeyapiv1alpha2.DefaultEtcdVersion, arch)
	kubeadm.Path = fmt.Sprintf("%s/kubeadm", filepath)
	kubelet.Path = fmt.Sprintf("%s/kubelet", filepath)
	kubectl.Path = fmt.Sprintf("%s/kubectl", filepath)
	kubecni.Path = fmt.Sprintf("%s/cni-plugins-linux-%s-%s.tgz", filepath, arch, kubekeyapiv1alpha2.DefaultCniVersion)
	helm.Path = fmt.Sprintf("%s/helm", filepath)
	docker.Path = fmt.Sprintf("%s/docker-%s.tgz", filepath, kubekeyapiv1alpha2.DefaultDockerVersion)
	crictl.Path = fmt.Sprintf("%s/crictl-%s-linux-%s.tar.gz", filepath, kubekeyapiv1alpha2.DefaultCrictlVersion, arch)

	if kkzone == "cn" {
		etcd.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		kubeadm.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", kubeadm.Version, kubeadm.Arch)
		kubelet.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", kubelet.Version, kubelet.Arch)
		kubectl.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", kubectl.Version, kubectl.Arch)
		kubecni.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, helm.Version)
		helm.GetCmd = kubeConf.Arg.DownloadCommand(helm.Path, helm.Url)
		docker.Url = fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(docker.Arch), docker.Version)
		crictl.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", kubekeyapiv1alpha2.DefaultCrictlVersion, kubekeyapiv1alpha2.DefaultCrictlVersion, arch)
	} else {
		etcd.Url = fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		kubeadm.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubeadm", kubeadm.Version, kubeadm.Arch)
		kubelet.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubelet", kubelet.Version, kubelet.Arch)
		kubectl.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl", kubectl.Version, kubectl.Arch)
		kubecni.Url = fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://get.helm.sh/helm-%s-linux-%s.tar.gz", helm.Version, helm.Arch)
		getCmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm-%s-linux-%s.tar.gz", filepath, helm.Version, helm.Arch), helm.Url)
		helm.GetCmd = fmt.Sprintf("%s && cd %s && tar -zxf helm-%s-linux-%s.tar.gz && mv linux-%s/helm . && rm -rf *linux-%s*", getCmd, filepath, helm.Version, helm.Arch, helm.Arch, helm.Arch)
		docker.Url = fmt.Sprintf("https://download.docker.com/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(docker.Arch), docker.Version)
		crictl.Url = fmt.Sprintf("https://github.com/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", kubekeyapiv1alpha2.DefaultCrictlVersion, kubekeyapiv1alpha2.DefaultCrictlVersion, arch)
	}

	kubeadm.GetCmd = kubeConf.Arg.DownloadCommand(kubeadm.Path, kubeadm.Url)
	kubelet.GetCmd = kubeConf.Arg.DownloadCommand(kubelet.Path, kubelet.Url)
	kubectl.GetCmd = kubeConf.Arg.DownloadCommand(kubectl.Path, kubectl.Url)
	kubecni.GetCmd = kubeConf.Arg.DownloadCommand(kubecni.Path, kubecni.Url)
	etcd.GetCmd = kubeConf.Arg.DownloadCommand(etcd.Path, etcd.Url)
	docker.GetCmd = kubeConf.Arg.DownloadCommand(docker.Path, docker.Url)
	crictl.GetCmd = kubeConf.Arg.DownloadCommand(crictl.Path, crictl.Url)

	binaries := []files.KubeBinary{kubeadm, kubelet, kubectl, helm, kubecni, docker, crictl, etcd}
	binariesMap := make(map[string]files.KubeBinary)
	for _, binary := range binaries {
		logger.Log.Messagef(common.LocalHost, "downloading %s ...", binary.Name)

		binariesMap[fmt.Sprintf("%s-%s", binary.Name, arch)] = binary
		if util.IsExist(binary.Path) {
			// download it again if it's incorrect
			if err := SHA256Check(binary); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
			} else {
				logger.Log.Messagef(common.LocalHost, "%s is existed", binary.Name)
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

	if kubeConf.Cluster.KubeSphere.Version == "v2.1.1" {
		logger.Log.Infoln(fmt.Sprintf("Downloading %s ...", "helm2"))
		if util.IsExist(fmt.Sprintf("%s/helm2", filepath)) == false {
			cmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm2", filepath), fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, "v2.16.9"))
			if output, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
				fmt.Println(string(output))
				return errors.Wrap(err, "Failed to download helm2 binary")
			}
		}
	}

	pipelineCache.Set(common.KubeBinaries, binariesMap)
	return nil
}

// SHA256Check is used to hash checks on downloaded binary. (sha256)
func SHA256Check(binary files.KubeBinary) error {
	output, err := sha256sum(binary.Path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", binary.Path))
	}

	if strings.TrimSpace(binary.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", binary.Name, binary.Version))
	}
	if output != binary.GetSha256() {
		return errors.New(fmt.Sprintf("SHA256 no match. %s not equal %s", binary.GetSha256(), output))
	}
	return nil
}

func sha256sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func KubernetesArtifactBinariesDownload(manifest *common.ArtifactManifest, path, arch string, pipelineCache *cache.Cache) error {
	kkzone := os.Getenv("KKZONE")

	m := manifest.Spec

	etcd := files.NewKubeBinary("etcd", arch, m.Components.ETCD.Version, path, kkzone, manifest.Arg.DownloadCommand)
	kubeadm := files.NewKubeBinary("kubeadm", arch, m.KubernetesDistribution.Version, path, kkzone, manifest.Arg.DownloadCommand)
	kubelet := files.NewKubeBinary("kubelet", arch, m.KubernetesDistribution.Version, path, kkzone, manifest.Arg.DownloadCommand)
	kubectl := files.NewKubeBinary("kubectl", arch, m.KubernetesDistribution.Version, path, kkzone, manifest.Arg.DownloadCommand)
	kubecni := files.NewKubeBinary("kubecni", arch, m.Components.CNI.Version, path, kkzone, manifest.Arg.DownloadCommand)
	helm := files.NewKubeBinary("helm", arch, m.Components.Helm.Version, path, kkzone, manifest.Arg.DownloadCommand)
	crictl := files.NewKubeBinary("crictl", arch, m.Components.Crictl.Version, path, kkzone, manifest.Arg.DownloadCommand)
	binaries := []files.KubeBinary{kubeadm, kubelet, kubectl, helm, kubecni, etcd}

	dockerArr := make([]files.KubeBinary, 0, 0)
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
			docker := files.NewKubeBinary("docker", arch, dockerVersion, path, kkzone, manifest.Arg.DownloadCommand)
			dockerArr = append(dockerArr, docker)
		}
	}

	binaries = append(binaries, dockerArr...)
	if m.Components.Crictl.Version != "" {
		binaries = append(binaries, crictl)
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
