package loadbalancer

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
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
