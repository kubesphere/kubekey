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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/loadbalancer/templates"
	"github.com/pkg/errors"
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

type CheckVIPAddress struct {
	common.KubeAction
}

func (c *CheckVIPAddress) Execute(runtime connector.Runtime) error {
	if c.KubeConf.Cluster.ControlPlaneEndpoint.Address == "" {
		return errors.New("VIP address is empty")
	} else {
		return nil
	}
}

type GetInterfaceName struct {
	common.KubeAction
}

func (g *GetInterfaceName) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if g.KubeConf.Cluster.ControlPlaneEndpoint.KubeVip.Mode == "BGP" {
		host.GetCache().Set("interface", "lo")
		return nil
	}
	cmd := fmt.Sprintf("ip route "+
		"| grep %s "+
		"| sed -e \"s/^.*dev.//\" -e \"s/.proto.*//\"", host.GetAddress())
	interfaceName, err := runtime.GetRunner().SudoCmd(cmd, false)
	if err != nil {
		return err
	}
	if interfaceName == "" {
		return errors.New("get interface failed")
	}
	// type: string
	host.GetCache().Set("interface", interfaceName)
	return nil
}

type GenerateKubevipManifest struct {
	common.KubeAction
}

func (g *GenerateKubevipManifest) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	interfaceName, ok := host.GetCache().GetMustString("interface")
	if !ok {
		return errors.New("get interface failed")
	}
	BGPMode := g.KubeConf.Cluster.ControlPlaneEndpoint.KubeVip.Mode == "BGP"
	hosts := runtime.GetHostsByRole(common.Master)
	var BGPPeersArr []string
	for _, value := range hosts {
		address := value.GetAddress()
		if address == host.GetAddress() {
			continue
		}
		BGPPeersArr = append(BGPPeersArr, fmt.Sprintf("%s:65000::false", address))
	}
	BGPPeers := strings.Join(BGPPeersArr, ",")
	templateAction := action.Template{
		Template: templates.KubevipManifest,
		Dst:      filepath.Join(common.KubeManifestDir, templates.KubevipManifest.Name()),
		Data: util.Data{
			"BGPMode":      BGPMode,
			"VipInterface": interfaceName,
			"BGPRouterID":  host.GetAddress(),
			"BGPPeers":     BGPPeers,
			"KubeVip":      g.KubeConf.Cluster.ControlPlaneEndpoint.Address,
			"KubevipImage": images.GetImage(runtime, g.KubeConf, "kubevip").ImageName(),
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
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

type CreateManifestsFolder struct {
	action.BaseAction
}

func (h *CreateManifestsFolder) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd("mkdir -p /var/lib/rancher/k3s/server/manifests/", false)
	if err != nil {
		return err
	}
	return nil
}

type GenerateK3sKubevipDaemonset struct {
	common.KubeAction
}

func (g *GenerateK3sKubevipDaemonset) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	interfaceName, ok := host.GetCache().GetMustString("interface")
	if !ok {
		return errors.New("get interface failed")
	}
	BGPMode := g.KubeConf.Cluster.ControlPlaneEndpoint.KubeVip.Mode == "BGP"
	hosts := runtime.GetHostsByRole(common.Master)
	var BGPPeersArr []string
	for _, value := range hosts {
		address := value.GetAddress()
		if address == host.GetAddress() {
			continue
		}
		BGPPeersArr = append(BGPPeersArr, fmt.Sprintf("%s:65000::false", address))
	}
	BGPPeers := strings.Join(BGPPeersArr, ",")
	templateAction := action.Template{
		Template: templates.K3sKubevipManifest,
		Dst:      filepath.Join("/var/lib/rancher/k3s/server/manifests/", templates.K3sKubevipManifest.Name()),
		Data: util.Data{
			"BGPMode":        BGPMode,
			"KubeVipVersion": images.GetImage(runtime, g.KubeConf, "kubevip").Tag,
			"VipInterface":   interfaceName,
			"BGPRouterID":    host.GetAddress(),
			"BGPPeers":       BGPPeers,
			"KubeVip":        g.KubeConf.Cluster.ControlPlaneEndpoint.Address,
			"KubevipImage":   images.GetImage(runtime, g.KubeConf, "kubevip").ImageName(),
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type DeleteVIP struct {
	common.KubeAction
}

func (g *DeleteVIP) Execute(runtime connector.Runtime) error {
	if g.KubeConf.Cluster.ControlPlaneEndpoint.KubeVip.Mode == "BGP" {
		cmd := fmt.Sprintf("ip addr del %s dev lo", g.KubeConf.Cluster.ControlPlaneEndpoint.Address)
		_, err := runtime.GetRunner().SudoCmd(cmd, false)
		if err != nil {
			return err
		}
	}
	return nil
}
