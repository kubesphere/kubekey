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
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

// Kubeadm info
const (
	KubeadmName          = "kubeadm"
	KubeadmID            = "kubeadm"
	KubeadmURLPathTmpl   = "/kubernetes-release/release/%s/bin/linux/%s/kubeadm"
	KubeadmURLPathTmplCN = "/release/%s/bin/linux/%s/kubeadm"
)

// Kubeadm is a Binary for kubeadm.
type Kubeadm struct {
	*Binary
}

// NewKubeadm returns a new Kubeadm.
func NewKubeadm(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Kubeadm, error) {
	fileName := KubeadmName
	file, err := NewFile(Params{
		SSHClient:      sshClient,
		RootFs:         rootFs,
		Type:           FileBinary,
		Name:           fileName,
		LocalFullPath:  filepath.Join(rootFs.ClusterRootFsDir(), fileName),
		RemoteFullPath: filepath.Join(BinDir, fileName),
	})
	if err != nil {
		return nil, err
	}

	u := parseURL(DefaultDownloadHostGoogle, fmt.Sprintf(KubeadmURLPathTmpl, version, arch))
	internal := checksum.NewChecksum(KubeadmID, version, arch)
	binary := NewBinary(BinaryParams{
		File:     file,
		ID:       KubeadmID,
		Version:  version,
		Arch:     arch,
		URL:      u,
		Checksum: internal,
	})

	return &Kubeadm{binary}, nil
}

// SetZone override Binary's SetZone method.
func (k *Kubeadm) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		k.SetHost(DefaultDownloadHostQingStor)
		k.SetPath(fmt.Sprintf(KubeadmURLPathTmplCN, k.version, k.arch))
	}
}
