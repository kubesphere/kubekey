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
	"github.com/kubesphere/kubekey/pkg/util"
)

// Docker info
const (
	DockerName           = "docker-%s.tgz"
	DockerID             = "docker"
	DockerURL            = "https://download.docker.com"
	DockerURLPathTmpl    = "/linux/static/stable/%s/docker-%s.tgz"
	DockerURLCN          = "https://mirrors.aliyun.com"
	DockerURLPathTmplCN  = "/docker-ce/linux/static/stable/%s/docker-%s.tgz"
	DockerDefaultVersion = "20.10.8"
)

// Docker is a Binary for docker.
type Docker struct {
	*Binary
}

// NewDocker returns a new Docker.
func NewDocker(sshClient ssh.Interface, rootFs rootfs.Interface, version, arch string) (*Docker, error) {
	fileName := fmt.Sprintf(DockerName, version)
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

	u := parseURL(DockerURL, fmt.Sprintf(DockerURLPathTmpl, util.ArchAlias(arch), version))
	binary := NewBinary(BinaryParams{
		File:    file,
		ID:      DockerID,
		Version: version,
		Arch:    arch,
		URL:     u,
	})

	return &Docker{binary}, nil
}

// SetZone override Binary's SetZone method.
func (d *Docker) SetZone(zone string) {
	if strings.EqualFold(zone, ZONE) {
		d.SetHost(DockerURLCN)
		d.SetPath(fmt.Sprintf(DockerURLPathTmplCN, util.ArchAlias(d.arch), d.version))
	}
}
