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

	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/rootfs"
)

// K3s info
const (
	K3sName          = "k3s"
	K3sID            = "k3s"
	K3sURLPathTmpl   = "/k3s-io/k3s/releases/download/%s+k3s1/k3s%s"
	K3sURLPathTmplCN = "/k3s/releases/download/%s+k3s1/linux/%s/k3s"
)

// K3s is a Binary for k3s.
type K3s struct {
	*Binary
}

// NewK3s returns a new K3s.
func NewK3s(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*K3s, error) {
	fileName := K3sName
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

	if arch == "amd64" {
		arch = ""
	} else {
		arch = "-" + arch
	}

	u := parseURL(DefaultDownloadHostGoogle, fmt.Sprintf(K3sURLPathTmpl, version, arch))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      K3sID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &K3s{binary}, nil
}

// SetZone override Binary's SetZone method.
func (k *K3s) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		k.SetHost(DefaultDownloadHostQingStor)
		k.SetPath(fmt.Sprintf(K3sURLPathTmplCN, k.version, k.arch))
	}
}
