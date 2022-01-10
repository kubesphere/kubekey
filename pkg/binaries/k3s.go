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
	"strings"
)

// K3sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K3sFilesDownloadHTTP(kubeConf *common.KubeConf, filepath, version, arch string, pipelineCache *cache.Cache) error {
	kkzone := os.Getenv("KKZONE")
	etcd := files.KubeBinary{Name: "etcd", Arch: arch, Version: kubekeyapiv1alpha2.DefaultEtcdVersion}
	k3s := files.KubeBinary{Name: "k3s", Arch: arch, Version: version}
	kubecni := files.KubeBinary{Name: "kubecni", Arch: arch, Version: kubekeyapiv1alpha2.DefaultCniVersion}
	helm := files.KubeBinary{Name: "helm", Arch: arch, Version: kubekeyapiv1alpha2.DefaultHelmVersion}

	etcd.Path = fmt.Sprintf("%s/etcd-%s-linux-%s.tar.gz", filepath, kubekeyapiv1alpha2.DefaultEtcdVersion, arch)
	k3s.Path = fmt.Sprintf("%s/k3s", filepath)
	kubecni.Path = fmt.Sprintf("%s/cni-plugins-linux-%s-%s.tgz", filepath, arch, kubekeyapiv1alpha2.DefaultCniVersion)
	helm.Path = fmt.Sprintf("%s/helm", filepath)

	if kkzone == "cn" {
		etcd.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		k3s.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/k3s/releases/download/%s+k3s1/linux/%s/k3s", k3s.Version, k3s.Arch)
		kubecni.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, helm.Version)
		helm.GetCmd = kubeConf.Arg.DownloadCommand(helm.Path, helm.Url)
	} else {
		etcd.Url = fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz", etcd.Version, etcd.Version, etcd.Arch)
		if k3s.Arch == kubekeyapiv1alpha2.DefaultArch {
			k3s.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s", k3s.Version)
		} else {
			k3s.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s-%s", k3s.Version, k3s.Arch)
		}
		kubecni.Url = fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
		helm.Url = fmt.Sprintf("https://get.helm.sh/helm-%s-linux-%s.tar.gz", helm.Version, helm.Arch)
		getCmd := kubeConf.Arg.DownloadCommand(fmt.Sprintf("%s/helm-%s-linux-%s.tar.gz", filepath, helm.Version, helm.Arch), helm.Url)
		helm.GetCmd = fmt.Sprintf("%s && cd %s && tar -zxf helm-%s-linux-%s.tar.gz && mv linux-%s/helm . && rm -rf *linux-%s*", getCmd, filepath, helm.Version, helm.Arch, helm.Arch, helm.Arch)
	}

	k3s.GetCmd = kubeConf.Arg.DownloadCommand(k3s.Path, k3s.Url)
	kubecni.GetCmd = kubeConf.Arg.DownloadCommand(kubecni.Path, kubecni.Url)
	etcd.GetCmd = kubeConf.Arg.DownloadCommand(etcd.Path, etcd.Url)

	binaries := []files.KubeBinary{k3s, helm, kubecni, etcd}
	binariesMap := make(map[string]files.KubeBinary)
	for _, binary := range binaries {
		logger.Log.Messagef(common.LocalHost, "downloading %s %s ...", arch, binary.Name)

		binariesMap[binary.Name] = binary
		if util.IsExist(binary.Path) {
			// download it again if it's incorrect
			if err := K3sSHA256Check(binary); err != nil {
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

			if err := K3sSHA256Check(binary); err != nil {
				if i == 1 {
					return err
				}
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path)).Run()
				continue
			}
			break
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}

// K3sSHA256Check is used to hash checks on downloaded binary. (sha256)
func K3sSHA256Check(binary files.KubeBinary) error {
	output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sha256sum %s", binary.Path)).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", binary.Path))
	}

	if strings.TrimSpace(binary.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", binary.Name, binary.Version))
	}
	if !strings.Contains(strings.TrimSpace(string(output)), binary.GetSha256()) {
		return errors.New(fmt.Sprintf("SHA256 no match. %s not in %s", binary.GetSha256(), strings.TrimSpace(string(output))))
	}
	return nil
}
