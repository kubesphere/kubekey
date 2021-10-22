package loadbalancer

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/loadbalancer/templates"
	"github.com/pkg/errors"
	"path/filepath"
	"strconv"
)

type GetChecksum struct {
	common.KubeAction
}

func (g *GetChecksum) Execute(runtime connector.Runtime) error {
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
			"HaproxyImage":    images.GetImage(runtime, g.KubeConf, "haproxy").ImageName(),
			"HealthCheckPort": 8081,
			"Checksum":        md5Str,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type UpdateK3s struct {
	common.KubeAction
}

func (u *UpdateK3s) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("sed -i 's#--server=.*\"#--server=https://127.0.0.1:%s\"#g' /etc/systemd/system/k3s.service", false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart k3s", false); err != nil {
		return err
	}
	return nil
}

type UpdateKubelet struct {
	common.KubeAction
}

func (u *UpdateKubelet) Execute(runtime connector.Runtime) error {
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

type UpdateKubeProxy struct {
	common.KubeAction
}

func (u *UpdateKubeProxy) Execute(runtime connector.Runtime) error {
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

type UpdateHosts struct {
	common.KubeAction
}

func (u *UpdateHosts) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sed -i 's#.* %s#127.0.0.1 %s#g' /etc/hosts",
		u.KubeConf.Cluster.ControlPlaneEndpoint.Domain, u.KubeConf.Cluster.ControlPlaneEndpoint.Domain), false); err != nil {
		return err
	}
	return nil
}

type GenerateK3sHaproxyManifest struct {
	common.KubeAction
}

func (g *GenerateK3sHaproxyManifest) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	md5Str, ok := host.GetCache().GetMustString("md5")
	if !ok {
		return errors.New("get haproxy config md5 sum by host label failed")
	}

	templateAction := action.Template{
		Template: templates.HaproxyManifest,
		Dst:      filepath.Join("/var/lib/rancher/k3s/agent/pod-manifests", templates.HaproxyManifest.Name()),
		Data: util.Data{
			"HaproxyImage":    images.GetImage(runtime, g.KubeConf, "haproxy").ImageName(),
			"HealthCheckPort": 8081,
			"Checksum":        md5Str,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}
