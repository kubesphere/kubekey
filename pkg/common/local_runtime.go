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

package common

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"os"
	"os/user"
	"runtime"
)

type LocalRuntime struct {
	connector.BaseRuntime
}

func NewLocalRuntime(debug, ingoreErr bool) (LocalRuntime, error) {
	var localRuntime LocalRuntime
	u, err := user.Current()
	if err != nil {
		return localRuntime, err
	}
	if u.Username != "root" {
		return localRuntime, fmt.Errorf("current user is %s. Please use root", u.Username)
	}

	name, err := os.Hostname()
	if err != nil {
		return localRuntime, err
	}
	base := connector.NewBaseRuntime(name, connector.NewDialer(), debug, ingoreErr)

	host := connector.NewHost()
	host.Name = name
	host.Address = util.LocalIP()
	host.InternalAddress = util.LocalIP()
	host.Port = 22
	host.User = u.Name
	host.Password = ""
	host.PrivateKeyPath = fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir)
	host.Arch = runtime.GOARCH
	host.SetRole(KubeKey)

	base.AppendHost(host)
	base.AppendRoleMap(host)

	local := LocalRuntime{base}
	return local, nil
}
