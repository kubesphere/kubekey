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

// Kubectl info
const (
	KubectlName          = "kubectl"
	KubectlID            = "kubectl"
	KubectlURLPathTmpl   = "/kubernetes-release/release/%s/bin/linux/%s/kubectl"
	KubectlURLPathTmplCN = "/release/%s/bin/linux/%s/kubectl"
)

// Kubectl is a Binary for kubectl.
type Kubectl struct {
	*Binary
}

// NewKubectl returns a new Kubectl.
func NewKubectl(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Kubectl, error) {
	fileName := KubectlName
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

	u := parseURL(DefaultDownloadHostGoogle, fmt.Sprintf(KubectlURLPathTmpl, version, arch))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      KubectlID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Kubectl{binary}, nil
}

// SetZone override Binary's SetZone method.
func (k *Kubectl) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		k.SetHost(DefaultDownloadHostQingStor)
		k.SetPath(fmt.Sprintf(KubectlURLPathTmplCN, k.version, k.arch))
	}
}
