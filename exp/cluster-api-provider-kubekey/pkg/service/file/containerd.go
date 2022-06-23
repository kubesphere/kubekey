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

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/file/checksum"
)

const (
	ContainerdName            = "containerd-%s-linux-%s.tar.gz"
	ContainerdID              = "containerd"
	ContainerdDownloadURLTmpl = "https://github.com/containerd/containerd/releases/download/v%s/containerd-%s-linux-%s.tar.gz"
)

func NewContainerd(file *File, version, arch string) (*Binary, error) {
	internal := checksum.NewChecksum(ContainerdID, version, arch)

	file.name = fmt.Sprintf(ContainerdName, version, arch)
	file.localFullPath = filepath.Join(file.name)
	file.remoteFullPath = filepath.Join(BinDir, file.name)

	return &Binary{
		file,
		ContainerdID,
		version,
		arch,
		fmt.Sprintf(ContainerdDownloadURLTmpl, version, version, arch),
		internal,
	}, nil
}
