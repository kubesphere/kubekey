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

type Debian struct {
	runtime connector.Runtime
	backup  bool
}

func NewDeb(runtime connector.Runtime) Interface {
	return &Debian{
		runtime: runtime,
	}
}

func (d *Debian) Backup() error {
	if _, err := d.runtime.GetRunner().SudoCmd("cp -r /etc/apt/sources.list.d /etc/apt/sources.list.d.kubekey.bak", false); err != nil {
		return err
	}
	d.backup = true
	return nil
}

func (d *Debian) IsAlreadyBackUp() bool {
	return d.backup
}

func (d *Debian) Add(path string) error {
	if !d.IsAlreadyBackUp() {
		return fmt.Errorf("linux repository must be backuped before")
	}

	if _, err := d.runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d/*", false); err != nil {
		return err
	}

	if _, err := d.runtime.GetRunner().SudoCmd(fmt.Sprintf("echo 'deb [trusted=yes]  file://%s   /' > /etc/apt/sources.list.d/kubekey.list", path),
		true); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Update() error {
	if _, err := d.runtime.GetRunner().SudoCmd("apt-get update && apt-get upgrade -y", true); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Install(pkg ...string) error {
	if len(pkg) == 0 {
		pkg = []string{"socat", "conntrack", "ipset", "ebtables", "chrony"}
	}

	str := strings.Join(pkg, " ")
	if _, err := d.runtime.GetRunner().SudoCmd(fmt.Sprintf("apt install -y %s", str), true); err != nil {
		return err
	}
	return nil
}

func (d *Debian) Reset() error {
	if _, err := d.runtime.GetRunner().SudoCmd("rm -f /etc/apt/sources.list.d/kubekey.list", false); err != nil {
		return err
	}

	if _, err := d.runtime.GetRunner().SudoCmd("cp -r /etc/apt/sources.list.d.kubekey.bak/* /etc/apt/sources.list.d/", false); err != nil {
		return err
	}

	if _, err := d.runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d.kubekey.bak", false); err != nil {
		return err
	}

	if err := d.Update(); err != nil {
		return err
	}
	return nil
}
