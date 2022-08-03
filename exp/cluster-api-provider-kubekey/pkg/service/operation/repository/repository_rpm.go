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
	"strings"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
)

type RedhatPackageManager struct {
	SSHClient ssh.Interface
}

func NewRPM(sshClient ssh.Interface) *RedhatPackageManager {
	return &RedhatPackageManager{
		SSHClient: sshClient,
	}
}

func (r *RedhatPackageManager) Update() error {
	if _, err := r.SSHClient.SudoCmd("yum clean all && yum makecache"); err != nil {
		return err
	}
	return nil
}

func (r *RedhatPackageManager) Install(pkg ...string) error {
	if len(pkg) == 0 {
		pkg = []string{"openssl", "socat", "conntrack", "ipset", "ebtables", "chrony", "ipvsadm"}
	}
	if _, err := r.SSHClient.SudoCmdf("apt install -y %s", strings.Join(pkg, " ")); err != nil {
		return err
	}
	return nil
}
