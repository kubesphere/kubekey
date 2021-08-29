package loadbalancer

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"strings"
)

type haproxyPreparatoryWork struct {
	common.KubeAction
}

func (h *haproxyPreparatoryWork) Execute(runtime connector.Runtime) error {
	if err := runtime.GetRunner().MkDir("/etc/kubekey/haproxy"); err != nil {
		return err
	}
	if err := runtime.GetRunner().Chmod("/etc/kubekey/haproxy", os.FileMode(0777)); err != nil {
		return err
	}
	return nil
}

type getChecksum struct {
	common.KubeAction
}

func (g *getChecksum) Execute(runtime connector.Runtime) error {
	md5Str, err := runtime.GetRunner().FileMd5("/etc/kubekey/haproxy/haproxy.cfg")
	if err != nil {
		return err
	}
	g.Cache.Set("md5", md5Str)
	return nil
}

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

type updateK3s struct {
	common.KubeAction
}

func (u *updateK3s) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("sed -i 's#--server=.*\"#--server=https://127.0.0.1:%s\"#g' /etc/systemd/system/k3s.service", false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart k3s", false); err != nil {
		return err
	}
	return nil
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

type updateKubelet struct {
	common.KubeAction
}

func (u *updateKubelet) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"sed -i 's#server:.*#server: https://127.0.0.1:%s#g' /etc/kubernetes/kubelet.conf",
		strconv.Itoa(u.KubeConf.Cluster.ControlPlaneEndpoint.Port)), false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart kubelet", false); err != nil {
		return err
	}
	return nil
}

type updateKubeproxyPrapre struct {
	common.KubePrepare
}

func (u *updateKubeproxyPrapre) PreCheck(runtime connector.Runtime) (bool, error) {
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

type updateKubeproxy struct {
	common.KubeAction
}

func (u *updateKubeproxy) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("set -o pipefail "+
		"&& /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf get configmap kube-proxy -n kube-system -o yaml "+
		"| sed 's#server:.*#server: https://127.0.0.1:%s#g' "+
		"| /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf replace -f -", false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf delete pod -n kube-system -l k8s-app=kube-proxy --force --grace-period=0", false); err != nil {
		return err
	}
	return nil
}

type updateHosts struct {
	common.KubeAction
}

func (u *updateHosts) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("sed -i 's#.* %s#127.0.0.1 %s#g' /etc/hosts", false); err != nil {
		return err
	}
	return nil
}
