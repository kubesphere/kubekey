package loadbalancer

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/images"
	"github.com/kubesphere/kubekey/pkg/pipelines/loadbalancer/templates"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strconv"
)

type haproxyPreparatoryWork struct {
	common.KubeAction
}

func (h *haproxyPreparatoryWork) Execute(runtime connector.Runtime) error {
	if err := runtime.GetRunner().MkDir(common.HaproxyDir); err != nil {
		return err
	}
	if err := runtime.GetRunner().Chmod(common.HaproxyDir, os.FileMode(0777)); err != nil {
		return err
	}
	return nil
}

type getChecksum struct {
	common.KubeAction
}

func (g *getChecksum) Execute(runtime connector.Runtime) error {
	md5Str, err := runtime.GetRunner().FileMd5(filepath.Join(common.HaproxyDir, "haproxy.cfg"))
	if err != nil {
		return err
	}
	host := runtime.RemoteHost()
	// type: string
	host.GetCache().Set("md5", md5Str)
	return nil
}

type GenerateHaproxyManifest struct {
	common.KubeAction
}

func (g *GenerateHaproxyManifest) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	md5Str, ok := host.GetCache().GetMustString("md5")
	if !ok {
		return errors.New("get haproxy config md5 sum by host label failed")
	}

	templateAction := action.Template{
		Template: templates.HaproxyManifest,
		Dst:      filepath.Join(common.KubeManifestDir, templates.HaproxyManifest.Name()),
		Data: util.Data{
			"HaproxyImage":    images.GetImage(g.Runtime, g.KubeConf, "haproxy").ImageName(),
			"HealthCheckPort": 8081,
			"Checksum":        md5Str,
		},
	}

	templateAction.Init(nil, nil, runtime)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
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

type updateKubeproxy struct {
	common.KubeAction
}

func (u *updateKubeproxy) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"set -o pipefail "+
			"&& /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf get configmap kube-proxy -n kube-system -o yaml "+
			"| sed 's#server:.*#server: https://127.0.0.1:%s#g' "+
			"| /usr/local/bin/kubectl --kubeconfig /etc/kubernetes/admin.conf replace -f -",
		strconv.Itoa(u.KubeConf.Cluster.ControlPlaneEndpoint.Port)), false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl "+
		"--kubeconfig /etc/kubernetes/admin.conf delete pod "+
		"-n kube-system "+
		"-l k8s-app=kube-proxy "+
		"--force "+
		"--grace-period=0", false); err != nil {
		return err
	}
	return nil
}

type updateHosts struct {
	common.KubeAction
}

func (u *updateHosts) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sed -i 's#.* %s#127.0.0.1 %s#g' /etc/hosts",
		u.KubeConf.Cluster.ControlPlaneEndpoint.Domain, u.KubeConf.Cluster.ControlPlaneEndpoint.Domain), false); err != nil {
		return err
	}
	return nil
}
