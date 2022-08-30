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

// kubelet info
const (
	KubeletName          = "kubelet"
	KubeletID            = "kubelet"
	KubeletURLPathTmpl   = "/kubernetes-release/release/%s/bin/linux/%s/kubelet"
	KubeletURLPathTmplCN = "/release/%s/bin/linux/%s/kubelet"
)

// Kubelet is a Binary for kubelet.
type Kubelet struct {
	*Binary
}

// NewKubelet returns a new Kubelet.
func NewKubelet(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Kubelet, error) {
	fileName := KubeletName
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

	u := parseURL(DefaultDownloadHostGoogle, fmt.Sprintf(KubeletURLPathTmpl, version, arch))
	internal := checksum.NewChecksum(KubeletID, version, arch)
	binary := NewBinary(BinaryParams{
		File:     file,
		ID:       KubeletID,
		Version:  version,
		Arch:     arch,
		URL:      u,
		Checksum: internal,
	})

	return &Kubelet{binary}, nil
}

// SetZone override Binary's SetZone method.
func (k *Kubelet) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		k.SetHost(DefaultDownloadHostQingStor)
		k.SetPath(fmt.Sprintf(KubeletURLPathTmplCN, k.version, k.arch))
	}
}
