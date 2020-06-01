/*
Copyright 2020 The KubeSphere Authors.

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

package preinstall

import (
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FilesDownloadHttp(mgr *manager.Manager, filepath, version string) error {

	kubeadm := files.KubeBinary{Name: "kubeadm", Arch: kubekeyapi.DefaultArch, Version: version}
	kubelet := files.KubeBinary{Name: "kubelet", Arch: kubekeyapi.DefaultArch, Version: version}
	kubectl := files.KubeBinary{Name: "kubectl", Arch: kubekeyapi.DefaultArch, Version: version}
	kubecni := files.KubeBinary{Name: "kubecni", Arch: kubekeyapi.DefaultArch, Version: kubekeyapi.DefaultCniVersion}
	helm := files.KubeBinary{Name: "helm", Arch: kubekeyapi.DefaultArch, Version: kubekeyapi.DefaultHelmVersion}

	kubeadm.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", kubeadm.Version, kubeadm.Arch)
	kubelet.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", kubelet.Version, kubelet.Arch)
	kubectl.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", kubectl.Version, kubectl.Arch)
	kubecni.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", kubecni.Version, kubecni.Arch, kubecni.Version)
	helm.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", helm.Arch, helm.Version)

	kubeadm.Path = fmt.Sprintf("%s/kubeadm", filepath)
	kubelet.Path = fmt.Sprintf("%s/kubelet", filepath)
	kubectl.Path = fmt.Sprintf("%s/kubectl", filepath)
	kubecni.Path = fmt.Sprintf("%s/cni-plugins-linux-%s-%s.tgz", filepath, kubekeyapi.DefaultArch, kubekeyapi.DefaultCniVersion)
	helm.Path = fmt.Sprintf("%s/helm", filepath)

	kubeadm.GetCmd = fmt.Sprintf("curl -o %s  %s", kubeadm.Path, kubeadm.Url)
	kubelet.GetCmd = fmt.Sprintf("curl -o %s  %s", kubelet.Path, kubelet.Url)
	kubectl.GetCmd = fmt.Sprintf("curl -o %s  %s", kubectl.Path, kubectl.Url)
	kubecni.GetCmd = fmt.Sprintf("curl -o %s  %s", kubecni.Path, kubecni.Url)
	helm.GetCmd = fmt.Sprintf("curl -o %s  %s", helm.Path, helm.Url)

	binaries := []files.KubeBinary{kubeadm, kubelet, kubectl, kubecni, helm}

	for _, binary := range binaries {
		mgr.Logger.Infoln(fmt.Sprintf("Downloading %s ...", binary.Name))
		if util.IsExist(binary.Path) == false {
			if output, err := exec.Command("/bin/sh", "-c", binary.GetCmd).CombinedOutput(); err != nil {
				fmt.Println(string(output))
				return errors.Wrap(err, "Failed to download kubeadm binary")
			}

			if err := SHA256Check(binary, version); err != nil {
				return err
			}
		} else {
			if err := SHA256Check(binary, version); err != nil {
				return err
			}
		}
	}

	return nil
}

func Prepare(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Downloading Installation Files")
	cfg := mgr.Cluster
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Faild to get current dir")
	}

	var kubeVersion string
	if cfg.Kubernetes.Version == "" {
		kubeVersion = kubekeyapi.DefaultKubeVersion
	} else {
		kubeVersion = cfg.Kubernetes.Version
	}

	filepath := fmt.Sprintf("%s/%s/%s", currentDir, kubekeyapi.DefaultPreDir, kubeVersion)
	if err := util.CreateDir(filepath); err != nil {
		return errors.Wrap(err, "Failed to create download target dir")
	}

	if err := FilesDownloadHttp(mgr, filepath, kubeVersion); err != nil {
		return err
	}
	return nil
}

func SHA256Check(binary files.KubeBinary, version string) error {
	output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sha256sum %s", binary.Path)).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", binary.Path))
	}

	if strings.TrimSpace(binary.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", version, version))
	} else {
		if !strings.Contains(strings.TrimSpace(string(output)), binary.GetSha256()) {
			return errors.New(fmt.Sprintf("SHA256 no match. %s not in %s", binary.GetSha256(), strings.TrimSpace(string(output))))
		}
	}
	return nil
}
