/*
 Copyright 2021 The KubeSphere Authors.

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
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"strings"
)

type RedhatPackageManager struct {
	runtime connector.Runtime
	backup  bool
}

func NewRPM(runtime connector.Runtime) Interface {
	return &RedhatPackageManager{
		runtime: runtime,
	}
}

func (r *RedhatPackageManager) Backup() error {
	if _, err := r.runtime.GetRunner().SudoCmd("mv /etc/yum.repos.d /etc/yum.repos.d.kubekey.bak", false); err != nil {
		return err
	}

	if _, err := r.runtime.GetRunner().SudoCmd("mkdir -p /etc/yum.repos.d", false); err != nil {
		return err
	}
	r.backup = true
	return nil
}

func (r *RedhatPackageManager) IsAlreadyBackUp() bool {
	return r.backup
}

func (r *RedhatPackageManager) Add(path string) error {
	if !r.IsAlreadyBackUp() {
		return fmt.Errorf("linux repository must be backuped before")
	}

	if _, err := r.runtime.GetRunner().SudoCmd("rm -rf /etc/yum.repos.d/*", false); err != nil {
		return err
	}

	content := fmt.Sprintf(`cat << EOF > /etc/yum.repos.d/CentOS-local.repo
[base-local]
name=CentOS7.6-local

baseurl=file://%s

enabled=1 

gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-CentOS-7
EOF
`, path)
	if _, err := r.runtime.GetRunner().SudoCmd(content, false); err != nil {
		return err
	}

	return nil
}

func (r *RedhatPackageManager) Update() error {
	if _, err := r.runtime.GetRunner().SudoCmd("yum clean all && yum makecache", true); err != nil {
		return err
	}
	return nil
}

func (r *RedhatPackageManager) Install(pkg ...string) error {
	if len(pkg) == 0 {
		pkg = []string{"openssl", "socat", "conntrack", "ipset", "ebtables", "chrony"}
	}

	str := strings.Join(pkg, " ")
	if _, err := r.runtime.GetRunner().SudoCmd(fmt.Sprintf("yum install -y %s", str), true); err != nil {
		return err
	}
	return nil
}

func (r *RedhatPackageManager) Reset() error {
	if _, err := r.runtime.GetRunner().SudoCmd("rm -rf /etc/yum.repos.d", false); err != nil {
		return err
	}

	if _, err := r.runtime.GetRunner().SudoCmd("mv /etc/yum.repos.d.kubekey.bak /etc/yum.repos.d", false); err != nil {
		return err
	}

	return nil
}
