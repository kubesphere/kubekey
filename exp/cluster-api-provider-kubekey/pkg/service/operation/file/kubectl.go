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
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file/checksum"
)

const (
	KubectlName            = "kubectl"
	KubectlID              = "kubectl"
	KubectlDownloadURLTmpl = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/%s/kubectl"
)

func NewKubectl(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Binary, error) {
	internal := checksum.NewChecksum(KubectlID, version, arch)

	fileName := KubectlName
	file, err := NewFile(FileParams{
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

	return &Binary{
		file,
		KubectlID,
		version,
		arch,
		fmt.Sprintf(KubectlDownloadURLTmpl, version, arch),
		internal,
	}, nil
}
