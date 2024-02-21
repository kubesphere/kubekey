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

package files

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/version"
)

const (
	kubeadm    = "kubeadm"
	kubelet    = "kubelet"
	kubectl    = "kubectl"
	kubecni    = "kubecni"
	etcd       = "etcd"
	helm       = "helm"
	amd64      = "amd64"
	arm64      = "arm64"
	k3s        = "k3s"
	k8e        = "k8e"
	docker     = "docker"
	cridockerd = "cri-dockerd"
	crictl     = "crictl"
	registry   = "registry"
	harbor     = "harbor"
	compose    = "compose"
	containerd = "containerd"
	runc       = "runc"
	calicoctl  = "calicoctl"
)

// KubeBinary Type field const
const (
	CNI        = "cni"
	CRICTL     = "crictl"
	DOCKER     = "docker"
	CRIDOCKERD = "cri-dockerd"
	ETCD       = "etcd"
	HELM       = "helm"
	KUBE       = "kube"
	REGISTRY   = "registry"
	CONTAINERD = "containerd"
	RUNC       = "runc"
)

var (
	// FileSha256 is a hash table the storage the checksum of the binary files. It is parsed from 'version/components.json'.
	FileSha256 = map[string]map[string]map[string]string{}
)

func init() {
	FileSha256, _ = version.ParseFilesSha256(version.Components)
}

type KubeBinary struct {
	Type     string
	ID       string
	FileName string
	Arch     string
	Version  string
	Url      string
	BaseDir  string
	Zone     string
	getCmd   func(path, url string) string
}

func NewKubeBinary(name, arch, version, prePath string, getCmd func(path, url string) string) *KubeBinary {
	component := new(KubeBinary)
	component.ID = name
	component.Arch = arch
	component.Version = version
	component.Zone = os.Getenv("KKZONE")
	component.getCmd = getCmd

	switch name {
	case etcd:
		component.Type = ETCD
		component.FileName = fmt.Sprintf("etcd-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/coreos/etcd/releases/download/%s/etcd-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf(
				"https://kubernetes-release.pek3b.qingstor.com/etcd/release/download/%s/etcd-%s-linux-%s.tar.gz",
				component.Version, component.Version, component.Arch)
		}
	case kubeadm:
		component.Type = KUBE
		component.FileName = kubeadm
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubeadm", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubeadm", version, arch)
		}
	case kubelet:
		component.Type = KUBE
		component.FileName = kubelet
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubelet", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubelet", version, arch)
		}
	case kubectl:
		component.Type = KUBE
		component.FileName = kubectl
		component.Url = fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/release/%s/bin/linux/%s/kubectl", version, arch)
		}
	case kubecni:
		component.Type = CNI
		component.FileName = fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", arch, version)
		component.Url = fmt.Sprintf("https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", version, arch, version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz", version, arch, version)
		}
	case helm:
		component.Type = HELM
		component.FileName = helm
		component.Url = fmt.Sprintf("https://get.helm.sh/helm-%s-linux-%s.tar.gz", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-helm.pek3b.qingstor.com/linux-%s/%s/helm", arch, version)
		}
	case docker:
		component.Type = DOCKER
		component.FileName = fmt.Sprintf("docker-%s.tgz", version)
		component.Url = fmt.Sprintf("https://download.docker.com/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(arch), version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://mirrors.aliyun.com/docker-ce/linux/static/stable/%s/docker-%s.tgz", util.ArchAlias(arch), version)
		}
	case cridockerd:
		component.Type = CRIDOCKERD
		component.FileName = fmt.Sprintf("cri-dockerd-%s.tgz", version)
		component.Url = fmt.Sprintf("https://github.com/Mirantis/cri-dockerd/releases/download/v%s/cri-dockerd-%s.%s.tgz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/cri-dockerd/releases/download/v%s/cri-dockerd-%s.%s.tgz", version, version, arch)
		}
	case crictl:
		component.Type = CRICTL
		component.FileName = fmt.Sprintf("crictl-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz", version, version, arch)
		}
	case k3s:
		component.Type = KUBE
		component.FileName = k3s
		component.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s", version)
		if arch == arm64 {
			component.Url = fmt.Sprintf("https://github.com/k3s-io/k3s/releases/download/%s+k3s1/k3s-%s", version, arch)
		}
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/k3s/releases/download/%s+k3s1/linux/%s/k3s", version, arch)
		}
	case k8e:
		component.Type = KUBE
		component.FileName = k8e
		component.Url = fmt.Sprintf("https://github.com/xiaods/k8e/releases/download/%s+k8e2/k8e", version)
		if arch == arm64 {
			component.Url = fmt.Sprintf("https://github.com/xiaods/k8e/releases/download/%s+k8e2/k8e-%s", version, arch)
		}
	case registry:
		component.Type = REGISTRY
		component.FileName = fmt.Sprintf("registry-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/kubesphere/kubekey/releases/download/v2.0.0-alpha.1/registry-%s-linux-%s.tar.gz", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/registry/%s/registry-%s-linux-%s.tar.gz", version, version, arch)
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case harbor:
		component.Type = REGISTRY
		component.FileName = fmt.Sprintf("harbor-offline-installer-%s.tgz", version)
		component.Url = fmt.Sprintf("https://github.com/goharbor/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", version, version)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/harbor/releases/download/%s/harbor-offline-installer-%s.tgz", version, version)
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case compose:
		component.Type = REGISTRY
		component.FileName = "docker-compose-linux-x86_64"
		component.Url = fmt.Sprintf("https://github.com/docker/compose/releases/download/%s/docker-compose-linux-%s", version, util.ArchAlias(arch))
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/docker/compose/releases/download/%s/docker-compose-linux-%s", version, util.ArchAlias(arch))
		}
		component.BaseDir = filepath.Join(prePath, component.Type, component.ID, component.Version, component.Arch)
	case containerd:
		component.Type = CONTAINERD
		component.FileName = fmt.Sprintf("containerd-%s-linux-%s.tar.gz", version, arch)
		component.Url = fmt.Sprintf("https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz", version, version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz", version, version, arch)
		}
	case runc:
		component.Type = RUNC
		component.FileName = fmt.Sprintf("runc.%s", arch)
		component.Url = fmt.Sprintf("https://github.com/opencontainers/runc/releases/download/%s/runc.%s", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/opencontainers/runc/releases/download/%s/runc.%s", version, arch)
		}
	case calicoctl:
		component.Type = CNI
		component.FileName = calicoctl
		component.Url = fmt.Sprintf("https://github.com/projectcalico/calico/releases/download/%s/calicoctl-linux-%s", version, arch)
		if component.Zone == "cn" {
			component.Url = fmt.Sprintf("https://kubernetes-release.pek3b.qingstor.com/projectcalico/calico/releases/download/%s/calicoctl-linux-%s", version, arch)
		}
	default:
		logger.Log.Fatalf("unsupported kube binaries %s", name)
	}

	if component.BaseDir == "" {
		component.BaseDir = filepath.Join(prePath, component.Type, component.Version, component.Arch)
	}

	return component
}

func (b *KubeBinary) CreateBaseDir() error {
	if err := util.CreateDir(b.BaseDir); err != nil {
		return err
	}
	return nil
}

func (b *KubeBinary) Path() string {
	return filepath.Join(b.BaseDir, b.FileName)
}

func (b *KubeBinary) GetCmd() string {
	cmd := b.getCmd(b.Path(), b.Url)

	if b.ID == helm && b.Zone != "cn" {
		get := b.getCmd(filepath.Join(b.BaseDir, fmt.Sprintf("helm-%s-linux-%s.tar.gz", b.Version, b.Arch)), b.Url)
		cmd = fmt.Sprintf("%s && cd %s && tar -zxf helm-%s-linux-%s.tar.gz && mv linux-%s/helm . && rm -rf *linux-%s*",
			get, b.BaseDir, b.Version, b.Arch, b.Arch, b.Arch)
	}
	return cmd
}

func (b *KubeBinary) GetSha256() string {
	s := FileSha256[b.ID][b.Arch][b.Version]
	return s
}

func (b *KubeBinary) Download() error {
	for i := 5; i > 0; i-- {
		cmd := exec.Command("/bin/sh", "-c", b.GetCmd())
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = cmd.Stdout

		if err = cmd.Start(); err != nil {
			return err
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
			if os.Getenv("KKZONE") != "cn" {
				logger.Log.Warningln("Having a problem with accessing https://storage.googleapis.com? You can try again after setting environment 'export KKZONE=cn'")
			}
			return err
		}

		if err := b.SHA256Check(); err != nil {
			if i == 1 {
				return err
			}
			path := b.Path()
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", path)).Run()
			continue
		}
		break
	}
	return nil
}

// SHA256Check is used to hash checks on downloaded binary. (sha256)
func (b *KubeBinary) SHA256Check() error {
	output, err := sha256sum(b.Path())
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to check SHA256 of %s", b.Path()))
	}

	if strings.TrimSpace(b.GetSha256()) == "" {
		return errors.New(fmt.Sprintf("No SHA256 found for %s. %s is not supported.", b.ID, b.Version))
	}
	if output != b.GetSha256() {
		return errors.New(fmt.Sprintf("SHA256 no match. %s not equal %s", b.GetSha256(), output))
	}
	return nil
}

func sha256sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
