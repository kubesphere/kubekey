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

package loadbalancer

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/pkg/errors"
	"strings"
)

type updateK3sPrepare struct {
	common.KubePrepare
}

func (u *updateK3sPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist("/etc/systemd/system/k3s.service")
	if err != nil {
		return false, err
	}

	if exist {
		if out, err := runtime.GetRunner().SudoCmd("sed -n '/--server=.*/p' /etc/systemd/system/k3s.service", false); err != nil {
			return false, err
		} else {
			if strings.Contains(strings.TrimSpace(out), LocalServer) {
				logger.Log.Debugf("do not restart kubelet, /etc/systemd/system/k3s.service content is %s", out)
				return false, nil
			}
		}
	} else {
		return false, errors.New("Failed to find /etc/systemd/system/k3s.service")
	}
	return true, nil
}

type updateKubeletPrepare struct {
	common.KubePrepare
}

func (u *updateKubeletPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	exist, err := runtime.GetRunner().FileExist("/etc/kubernetes/kubelet.conf")
	if err != nil {
		return false, err
	}

	if exist {
		if out, err := runtime.GetRunner().SudoCmd("sed -n '/server:.*/p' /etc/kubernetes/kubelet.conf", true); err != nil {
			return false, err
		} else {
			if strings.Contains(strings.TrimSpace(out), LocalServer) {
				logger.Log.Debugf("do not restart kubelet, /etc/kubernetes/kubelet.conf content is %s", out)
				return false, nil
			}
		}
	} else {
		return false, errors.New("Failed to find /etc/kubernetes/kubelet.conf")
	}
	return true, nil
}

type updateKubeProxyPrapre struct {
	common.KubePrepare
}

func (u *updateKubeProxyPrapre) PreCheck(runtime connector.Runtime) (bool, error) {
	if out, err := runtime.GetRunner().SudoCmd(
		"set -o pipefail && /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf get configmap kube-proxy -n kube-system -o yaml "+
			"| sed -n '/server:.*/p'", false); err != nil {
		return false, err
	} else {
		if strings.Contains(strings.TrimSpace(out), LocalServer) {
			logger.Log.Debugf("do not restart kube-proxy, configmap kube-proxy content is %s", out)
			return false, nil
		}
	}
	return true, nil
}
