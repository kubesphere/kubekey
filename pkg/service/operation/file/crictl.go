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

	"github.com/kubesphere/kubekey/v3/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/v3/pkg/rootfs"
)

// Crictl info
const (
	CrictlName          = "crictl-%s-linux-%s.tar.gz"
	CrictlID            = "crictl"
	CrictlURLPathTmpl   = "/kubernetes-sigs/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz"
	CrictlURLPathTmplCN = "/cri-tools/releases/download/%s/crictl-%s-linux-%s.tar.gz"
)

// Crictl is a Binary for crictl.
type Crictl struct {
	*Binary
}

// NewCrictl returns a new Crictl.
func NewCrictl(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Crictl, error) {
	fileName := fmt.Sprintf(CrictlName, version, arch)
	file, err := NewFile(Params{
		SSHClient:      sshClient,
		RootFs:         rootFs,
		Type:           FileBinary,
		Name:           fileName,
		LocalFullPath:  filepath.Join(rootFs.ClusterRootFsDir(), CrictlID, version, arch, fileName),
		RemoteFullPath: filepath.Join(BinDir, fileName),
	})
	if err != nil {
		return nil, err
	}

	u := parseURL(DefaultDownloadHost, fmt.Sprintf(CrictlURLPathTmpl, version, version, arch))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      CrictlID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Crictl{binary}, nil
}

// SetZone override Binary's SetZone method.
func (c *Crictl) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		c.SetHost(DefaultDownloadHostQingStor)
		c.SetPath(fmt.Sprintf(CrictlURLPathTmplCN, c.version, c.version, c.arch))
	}
}
