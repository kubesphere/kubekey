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

package repository

import (
	"fmt"
	"strings"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
)

// RedhatPackageManager is a repository manager implementation for Redhat, Centos.
type RedhatPackageManager struct {
	SSHClient ssh.Interface
}

// NewRPM returns a new RedhatPackageManager.
func NewRPM(sshClient ssh.Interface) *RedhatPackageManager {
	return &RedhatPackageManager{
		SSHClient: sshClient,
	}
}

// Add adds a local repository using the iso file.
func (r *RedhatPackageManager) Add(path string) error {
	content := fmt.Sprintf(`cat << EOF > /etc/yum.repos.d/kubekey.repo
[base-local]
name=KubeKey-local
baseurl=file://%s
enabled=1
gpgcheck=0
EOF
`, path)

	if _, err := r.SSHClient.SudoCmd(content); err != nil {
		return err
	}
	return nil
}

// Update updates the repository cache.
func (r *RedhatPackageManager) Update() error {
	if _, err := r.SSHClient.SudoCmd("yum clean all && yum makecache"); err != nil {
		return err
	}
	return nil
}

// Install installs common packages.
func (r *RedhatPackageManager) Install(pkg ...string) error {
	if len(pkg) == 0 {
		pkg = []string{"openssl", "socat", "conntrack", "ipset", "ebtables", "chrony", "ipvsadm"}
	}
	if _, err := r.SSHClient.SudoCmdf("yum install -y %s", strings.Join(pkg, " ")); err != nil {
		return err
	}
	return nil
}
