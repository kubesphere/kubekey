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

package file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/rootfs"
)

// Kubecni info
const (
	KubecniName           = "cni-plugins-linux-%s-%s.tgz"
	KubecniID             = "kubecni"
	KubecniURLPathTmpl    = "/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"
	KubecniURLCN          = "https://containernetworking.pek3b.qingstor.com"
	KubecniURLPathTmplCN  = "/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"
	KubecniDefaultVersion = "v0.9.1"
)

// Kubecni is a Binary for kubecni.
type Kubecni struct {
	*Binary
}

// NewKubecni returns a new Kubecni.
func NewKubecni(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Kubecni, error) {
	fileName := fmt.Sprintf(KubecniName, arch, version)
	file, err := NewFile(Params{
		SSHClient:      sshClient,
		RootFs:         rootFs,
		Type:           FileBinary,
		Name:           fileName,
		LocalFullPath:  filepath.Join(rootFs.ClusterRootFsDir(), fileName),
		RemoteFullPath: filepath.Join(OptCniBinDir, fileName),
	})
	if err != nil {
		return nil, err
	}

	u := parseURL(DefaultDownloadHost, fmt.Sprintf(KubecniURLPathTmpl, version, arch, version))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      KubecniID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Kubecni{binary}, nil
}

// SetZone override Binary's SetZone method.
func (k *Kubecni) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		k.SetHost(KubecniURLCN)
		k.SetPath(fmt.Sprintf(KubecniURLPathTmplCN, k.version, k.arch, k.version))
	}
}
