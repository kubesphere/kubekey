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

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/rootfs"
)

// runc info
const (
	RuncName           = "runc.%s"
	RuncID             = "runc"
	RuncURLPathTmpl    = "/opencontainers/runc/releases/download/%s/runc.%s"
	RuncDefaultVersion = "v1.1.1"
)

// Runc is a Binary for runc.
type Runc struct {
	*Binary
}

// NewRunc returns a new Runc.
func NewRunc(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Runc, error) {
	fileName := fmt.Sprintf(RuncName, arch)
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

	u := parseURL(DefaultDownloadHost, fmt.Sprintf(RuncURLPathTmpl, version, arch))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      RuncID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Runc{binary}, nil
}
