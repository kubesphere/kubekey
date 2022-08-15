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
	KubecniName            = "cni-plugins-linux-%s-%s.tgz"
	KubecniID              = "kubecni"
	KubecniDownloadURLTmpl = "https://github.com/containernetworking/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"
	//KubecniDownloadURLTmpl = "https://containernetworking.pek3b.qingstor.com/plugins/releases/download/%s/cni-plugins-linux-%s-%s.tgz"
	KubecniDefaultVersion = "v0.9.1"
)

func NewKubecni(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Binary, error) {
	internal := checksum.NewChecksum(KubecniID, version, arch)

	fileName := fmt.Sprintf(KubecniName, arch, version)
	file, err := NewFile(FileParams{
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

	return &Binary{
		file,
		KubecniID,
		version,
		arch,
		fmt.Sprintf(KubecniDownloadURLTmpl, version, arch, version),
		internal,
	}, nil
}
