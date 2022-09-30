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

	"github.com/kubesphere/kubekey/pkg/clients/ssh"
)

// Debian is a repository manager implementation for Debian.
type Debian struct {
	SSHClient ssh.Interface
}

// NewDeb returns a new Debian repository manager.
func NewDeb(sshClient ssh.Interface) *Debian {
	return &Debian{
		SSHClient: sshClient,
	}
}

// Add adds a local repository using the iso file.
func (d *Debian) Add(path string) error {
	if _, err := d.SSHClient.SudoCmd(
		fmt.Sprintf("echo 'deb [trusted=yes]  file://%s   ./' "+
			"| sudo tee /etc/apt/sources.list.d/kubekey.list > /dev/null", path)); err != nil {
		return err
	}
	return nil
}

// Update updates the repository cache.
func (d *Debian) Update() error {
	if _, err := d.SSHClient.Cmd("sudo apt-get update"); err != nil {
		return err
	}
	return nil
}

// Install installs common packages.
func (d *Debian) Install(pkg ...string) error {
	if len(pkg) == 0 {
		pkg = []string{"socat", "conntrack", "ipset", "ebtables", "chrony", "ipvsadm"}
	}
	if _, err := d.SSHClient.SudoCmdf("apt install -y %s", strings.Join(pkg, " ")); err != nil {
		return err
	}
	return nil
}
