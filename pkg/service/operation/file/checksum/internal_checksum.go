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

package checksum

import (
	"github.com/kubesphere/kubekey/version"
	"github.com/pkg/errors"
)

const (
	kubeadm    = "kubeadm"
	kubelet    = "kubelet"
	kubectl    = "kubectl"
	kubecni    = "kubecni"
	etcd       = "etcd"
	helm       = "helm"
	amd64      = "amd64"
	arm64      = "arm64"
	k3s        = "k3s"
	docker     = "docker"
	cridockerd = "cri-dockerd"
	crictl     = "crictl"
	registry   = "registry"
	harbor     = "harbor"
	compose    = "compose"
	containerd = "containerd"
	runc       = "runc"
)

var (
	// FileSha256 is a hash table the storage the checksum of the binary files. It is parsed from 'version/components.json'.
	FileSha256 = map[string]map[string]map[string]string{}
)

func init() {
	FileSha256, _ = version.ParseFilesSha256(version.Components)
}

// InternalChecksum is the internal checksum implementation.
type InternalChecksum struct {
	ID      string
	Version string
	Arch    string
	value   string
}

// NewInternalChecksum returns a new internal checksum implementation given the binary information.
func NewInternalChecksum(id, version, arch string) *InternalChecksum {
	return &InternalChecksum{
		ID:      id,
		Version: version,
		Arch:    arch,
	}
}

// Get gets the internal checksum.
func (i *InternalChecksum) Get() error {
	value, ok := FileSha256[i.ID][i.Arch][i.Version]
	if !ok {
		return errors.New("unsupported version")
	}
	i.value = value
	return nil
}

// Value returns the internal checksum value.
func (i *InternalChecksum) Value() string {
	return i.value
}
