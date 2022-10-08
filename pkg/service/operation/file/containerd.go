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

	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/rootfs"
)

// Containerd info
const (
	ContainerdName        = "containerd-%s-linux-%s.tar.gz"
	ContainerdID          = "containerd"
	ContainerdURLPathTmpl = "/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz"
)

// Containerd is a Binary for containerd.
type Containerd struct {
	*Binary
}

// NewContainerd returns a new Containerd.
func NewContainerd(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Containerd, error) {
	fileName := fmt.Sprintf(ContainerdName, version, arch)
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

	u := parseURL(DefaultDownloadHost, fmt.Sprintf(ContainerdURLPathTmpl, version, version, arch))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      ContainerdID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Containerd{binary}, nil
}
