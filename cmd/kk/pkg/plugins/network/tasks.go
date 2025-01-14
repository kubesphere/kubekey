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

package network

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/network/templates"
)

//go:embed cilium-1.11.7.tgz hybridnet-0.6.6.tgz templates/calico.tmpl

var f embed.FS

type ReleaseCiliumChart struct {
	common.KubeAction
}

func (r *ReleaseCiliumChart) Execute(runtime connector.Runtime) error {
	fs, err := os.Create(fmt.Sprintf("%s/cilium.tgz", runtime.GetWorkDir()))
	if err != nil {
		return err
	}
	chartFile, err := f.Open("cilium-1.11.7.tgz")
	if err != nil {
		return err
	}
	defer chartFile.Close()

	_, err = io.Copy(fs, chartFile)
	if err != nil {
		return err
	}

	return nil
}

type SyncCiliumChart struct {
	common.KubeAction
}

func (s *SyncCiliumChart) Execute(runtime connector.Runtime) error {
	src := filepath.Join(runtime.GetWorkDir(), "cilium.tgz")
	dst := filepath.Join(common.TmpDir, "cilium.tgz")
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync cilium chart failed"))
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mv %s/cilium.tgz /etc/kubernetes", common.TmpDir), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync cilium chart failed")
	}
	return nil
}

type DeployCilium struct {
	common.KubeAction
}

func (d *DeployCilium) Execute(runtime connector.Runtime) error {
	ciliumImage := images.GetImage(runtime, d.KubeConf, "cilium").ImageName()
	ciliumOperatorImage := images.GetImage(runtime, d.KubeConf, "cilium-operator-generic").ImageName()

	cmd := fmt.Sprintf("/usr/local/bin/helm upgrade --install cilium /etc/kubernetes/cilium.tgz --namespace kube-system "+
		"--set operator.image.override=%s "+
		"--set operator.replicas=1 "+
		"--set image.override=%s "+
		"--set ipam.operator.clusterPoolIPv4PodCIDR=%s", ciliumOperatorImage, ciliumImage, d.KubeConf.Cluster.Network.KubePodsCIDR)

	if d.KubeConf.Cluster.Kubernetes.DisableKubeProxy {
		cmd = fmt.Sprintf("%s --set kubeProxyReplacement=strict --set k8sServiceHost=%s --set k8sServicePort=%d", cmd, d.KubeConf.Cluster.ControlPlaneEndpoint.Address, d.KubeConf.Cluster.ControlPlaneEndpoint.Port)
	}

	if _, err := runtime.GetRunner().SudoCmd(cmd, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy cilium failed")
	}
	return nil
}

type DeployNetworkPlugin struct {
	common.KubeAction
}

func (d *DeployNetworkPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/network-plugin.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy network plugin failed")
	}
	return nil
}

type DeployKubeovnPlugin struct {
	common.KubeAction
}

func (d *DeployKubeovnPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/kube-ovn-crd.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy kube-ovn-crd.yaml failed")
	}
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/ovn.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy ovn.yaml failed")
	}
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/kube-ovn.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy kube-ovn.yaml failed")
	}
	return nil
}

type DeployNetworkMultusPlugin struct {
	common.KubeAction
}

func (d *DeployNetworkMultusPlugin) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl apply -f /etc/kubernetes/multus-network-plugin.yaml --force", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy multus network plugin failed")
	}
	return nil
}

type LabelNode struct {
	common.KubeAction
}

func (l *LabelNode) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("/usr/local/bin/kubectl label no -l%s kube-ovn/role=master --overwrite",
			l.KubeConf.Cluster.Network.Kubeovn.Label),
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "label kubeovn/role=master in master node failed")
	}
	return nil
}

type GenerateSSL struct {
	common.KubeAction
}

func (g *GenerateSSL) Execute(runtime connector.Runtime) error {
	if exist, err := runtime.GetRunner().SudoCmd(
		"/usr/local/bin/kubectl get secret -n kube-system kube-ovn-tls --ignore-not-found",
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "find ovn secret failed")
	} else if exist != "" {
		return nil
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("docker run --rm -v %s:/etc/ovn %s bash generate-ssl.sh",
			runtime.GetWorkDir(), images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName()),
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "generate ovn secret failed")
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("/usr/local/bin/kubectl create secret generic -n kube-system kube-ovn-tls "+
			"--from-file=cacert=%s/cacert.pem "+
			"--from-file=cert=%s/ovn-cert.pem "+
			"--from-file=key=%s/ovn-privkey.pem",
			runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir()),
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "create ovn secret failed")
	}

	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("rm -rf %s/cacert.pem %s/ovn-cert.pem %s/ovn-privkey.pem %s/ovn-req.pem",
			runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir(), runtime.GetWorkDir()),
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "remove generated ovn secret file failed")
	}

	return nil
}

type GenerateKubeOVN struct {
	common.KubeAction
}

func (g *GenerateKubeOVN) Execute(runtime connector.Runtime) error {
	address, err := runtime.GetRunner().Cmd(
		"/usr/local/bin/kubectl get no -lkube-ovn/role=master --no-headers -o wide | awk '{print $6}' | tr \\\\n ','",
		true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get kube-ovn label node address failed")
	}

	count, err := runtime.GetRunner().Cmd(
		fmt.Sprintf("/usr/local/bin/kubectl get no -l%s --no-headers -o wide | wc -l | sed 's/ //g'",
			g.KubeConf.Cluster.Network.Kubeovn.Label), true)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "count kube-ovn label nodes num failed")
	}

	if count == "0" {
		return fmt.Errorf("no node with label: %s", g.KubeConf.Cluster.Network.Kubeovn.Label)
	}

	templateAction := action.Template{
		Template: templates.KubeOvnCrd,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOvnCrd.Name()),
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	templateAction = action.Template{
		Template: templates.OVN,
		Dst:      filepath.Join(common.KubeConfigDir, templates.OVN.Name()),
		Data: util.Data{
			"Address":               address,
			"Count":                 count,
			"KubeovnImage":          images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"TunnelType":            g.KubeConf.Cluster.Network.Kubeovn.TunnelType,
			"DpdkMode":              g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkMode,
			"DpdkVersion":           g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkVersion,
			"OvnVersion":            v1alpha2.DefaultKubeovnVersion,
			"EnableSSL":             g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"HwOffload":             g.KubeConf.Cluster.Network.Kubeovn.OvsOvn.HwOffload,
			"SvcYamlIpfamilypolicy": g.KubeConf.Cluster.Network.Kubeovn.SvcYamlIpfamilypolicy,
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	templateAction = action.Template{
		Template: templates.KubeOvn,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOvn.Name()),
		Data: util.Data{
			"Address":               address,
			"Count":                 count,
			"KubeovnImage":          images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"PodCIDR":               g.KubeConf.Cluster.Network.KubePodsCIDR,
			"SvcCIDR":               g.KubeConf.Cluster.Network.KubeServiceCIDR,
			"JoinCIDR":              g.KubeConf.Cluster.Network.Kubeovn.JoinCIDR,
			"PodGateway":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.PodGateway,
			"CheckGateway":          g.KubeConf.Cluster.Network.Kubeovn.KubeovnCheckGateway(),
			"LogicalGateway":        g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.LogicalGateway,
			"PingExternalAddress":   g.KubeConf.Cluster.Network.Kubeovn.KubeOvnPinger.PingerExternalAddress,
			"PingExternalDNS":       g.KubeConf.Cluster.Network.Kubeovn.KubeOvnPinger.PingerExternalDomain,
			"NetworkType":           g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.NetworkType,
			"TunnelType":            g.KubeConf.Cluster.Network.Kubeovn.TunnelType,
			"ExcludeIps":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.ExcludeIps,
			"PodNicType":            g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.PodNicType,
			"VlanID":                g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.VlanID,
			"VlanInterfaceName":     g.KubeConf.Cluster.Network.Kubeovn.KubeOvnController.VlanInterfaceName,
			"Iface":                 g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.Iface,
			"EnableSSL":             g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"EnableMirror":          g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.EnableMirror,
			"EnableLB":              g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableLB(),
			"EnableNP":              g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableNP(),
			"EnableEipSnat":         g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableEipSnat(),
			"EnableExternalVPC":     g.KubeConf.Cluster.Network.Kubeovn.KubeovnEnableExternalVPC(),
			"SvcYamlIpfamilypolicy": g.KubeConf.Cluster.Network.Kubeovn.SvcYamlIpfamilypolicy,
			"DpdkTunnelIface":       g.KubeConf.Cluster.Network.Kubeovn.Dpdk.DpdkTunnelIface,
			"CNIConfigPriority":     g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.CNIConfigPriority,
			"Modules":               g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.Modules,
			"RPMs":                  g.KubeConf.Cluster.Network.Kubeovn.KubeOvnCni.RPMs,
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	return nil
}

type ChmodKubectlKo struct {
	common.KubeAction
}

func (c *ChmodKubectlKo) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(
		fmt.Sprintf("chmod +x %s", filepath.Join(common.BinDir, templates.KubectlKo.Name())), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod +x kubectl-ko failed")
	}
	return nil
}

// ReleaseHybridnetChart is used to release hybridnet chart to local path
type ReleaseHybridnetChart struct {
	common.KubeAction
}

func (r *ReleaseHybridnetChart) Execute(runtime connector.Runtime) error {
	fs, err := os.Create(fmt.Sprintf("%s/hybridnet.tgz", runtime.GetWorkDir()))
	if err != nil {
		return err
	}
	chartFile, err := f.Open("hybridnet-0.6.6.tgz")
	if err != nil {
		return err
	}
	defer chartFile.Close()

	_, err = io.Copy(fs, chartFile)
	if err != nil {
		return err
	}

	return nil
}

// SyncHybridnetChart is used to sync hybridnet chart to contronplane
type SyncHybridnetChart struct {
	common.KubeAction
}

func (s *SyncHybridnetChart) Execute(runtime connector.Runtime) error {
	src := filepath.Join(runtime.GetWorkDir(), "hybridnet.tgz")
	dst := filepath.Join(common.TmpDir, "hybridnet.tgz")
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("sync hybridnet chart failed"))
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mv %s/hybridnet.tgz /etc/kubernetes", common.TmpDir), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync hybrident chart failed")
	}
	return nil
}

type DeployHybridnet struct {
	common.KubeAction
}

func (d *DeployHybridnet) Execute(runtime connector.Runtime) error {

	cmd := fmt.Sprintf("/usr/local/bin/helm upgrade --install hybridnet /etc/kubernetes/hybridnet.tgz --namespace kube-system "+
		"--set images.hybridnet.image=%s/%s "+
		"--set images.hybridnet.tag=%s "+
		"--set images.registryURL=%s ",
		images.GetImage(runtime, d.KubeConf, "hybridnet").ImageNamespace(),
		images.GetImage(runtime, d.KubeConf, "hybridnet").Repo,
		images.GetImage(runtime, d.KubeConf, "hybridnet").Tag,
		images.GetImage(runtime, d.KubeConf, "hybridnet").ImageRegistryAddr(),
	)

	if d.KubeConf.Cluster.Network.Hybridnet.EnableInit() {
		cmd = fmt.Sprintf("%s --set init.cidr=%s", cmd, d.KubeConf.Cluster.Network.KubePodsCIDR)
	} else {
		cmd = fmt.Sprintf("%s --set init=null", cmd)
	}

	if !d.KubeConf.Cluster.Network.Hybridnet.NetworkPolicy() {
		cmd = fmt.Sprintf("%s --set daemon.enableNetworkPolicy=false", cmd)
	}

	if d.KubeConf.Cluster.Network.Hybridnet.DefaultNetworkType != "" {
		cmd = fmt.Sprintf("%s --set defaultNetworkType=%s", cmd, d.KubeConf.Cluster.Network.Hybridnet.DefaultNetworkType)
	}

	if d.KubeConf.Cluster.Network.Hybridnet.PreferBGPInterfaces != "" {
		cmd = fmt.Sprintf("%s --set daemon.preferBGPInterfaces=%s", cmd, d.KubeConf.Cluster.Network.Hybridnet.PreferBGPInterfaces)
	}

	if d.KubeConf.Cluster.Network.Hybridnet.PreferVlanInterfaces != "" {
		cmd = fmt.Sprintf("%s --set daemon.preferVlanInterfaces=%s", cmd, d.KubeConf.Cluster.Network.Hybridnet.PreferVlanInterfaces)
	}

	if d.KubeConf.Cluster.Network.Hybridnet.PreferVxlanInterfaces != "" {
		cmd = fmt.Sprintf("%s --set daemon.preferVxlanInterfaces=%s", cmd, d.KubeConf.Cluster.Network.Hybridnet.PreferVxlanInterfaces)
	}

	if _, err := runtime.GetRunner().SudoCmd(cmd, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "deploy hybridnet failed")
	}

	if len(d.KubeConf.Cluster.Network.Hybridnet.Networks) != 0 {
		templateAction := action.Template{
			Template: templates.HybridnetNetworks,
			Dst:      filepath.Join(common.KubeConfigDir, templates.HybridnetNetworks.Name()),
			Data: util.Data{
				"Networks": d.KubeConf.Cluster.Network.Hybridnet.Networks,
			},
		}

		templateAction.Init(nil, nil)
		if err := templateAction.Execute(runtime); err != nil {
			return err
		}

		for i := 0; i < 30; i++ {
			fmt.Println("Waiting for hybridnet webhook running ... ", i+1)
			time.Sleep(10 * time.Second)
			output, _ := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get pod -n kube-system -l  app=hybridnet,component=webhook | grep Running", false)
			if strings.Contains(output, "1/1") {
				time.Sleep(50 * time.Second)
				break
			}
		}

		if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/hybridnet-networks.yaml", true); err != nil {
			return errors.Wrap(errors.WithStack(err), "apply hybridnet networks failed")
		}
	}
	return nil
}

type GenerateCalicoManifests struct {
	common.KubeAction
}

func (g *GenerateCalicoManifests) Execute(runtime connector.Runtime) error {
	calicoContent, err := f.ReadFile("templates/calico.tmpl")
	if err != nil {
		return err
	}
	calico := template.Must(template.New("network-plugin.yaml").Funcs(utils.FuncMap).Parse(string(calicoContent)))

	IPv6Support := false
	kubePodsV6CIDR := ""
	kubePodsCIDR := strings.Split(g.KubeConf.Cluster.Network.KubePodsCIDR, ",")
	if len(kubePodsCIDR) == 2 {
		IPv6Support = true
		kubePodsV6CIDR = kubePodsCIDR[1]
	}

	templateAction := action.Template{
		Template: calico,
		Dst:      filepath.Join(common.KubeConfigDir, calico.Name()),
		Data: util.Data{
			"KubePodsV4CIDR":          strings.Split(g.KubeConf.Cluster.Network.KubePodsCIDR, ",")[0],
			"KubePodsV6CIDR":          kubePodsV6CIDR,
			"CalicoCniImage":          images.GetImage(runtime, g.KubeConf, "calico-cni").ImageName(),
			"CalicoNodeImage":         images.GetImage(runtime, g.KubeConf, "calico-node").ImageName(),
			"CalicoFlexvolImage":      images.GetImage(runtime, g.KubeConf, "calico-flexvol").ImageName(),
			"CalicoControllersImage":  images.GetImage(runtime, g.KubeConf, "calico-kube-controllers").ImageName(),
			"CalicoTyphaImage":        images.GetImage(runtime, g.KubeConf, "calico-typha").ImageName(),
			"TyphaEnabled":            len(runtime.GetHostsByRole(common.K8s)) > 50 || g.KubeConf.Cluster.Network.Calico.EnableTypha(),
			"VethMTU":                 g.KubeConf.Cluster.Network.Calico.VethMTU,
			"NodeCidrMaskSize":        g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
			"IPIPMode":                g.KubeConf.Cluster.Network.Calico.IPIPMode,
			"VXLANMode":               g.KubeConf.Cluster.Network.Calico.VXLANMode,
			"ConatinerManagerIsIsula": g.KubeConf.Cluster.Kubernetes.ContainerManager == "isula",
			"IPV4POOLNATOUTGOING":     g.KubeConf.Cluster.Network.Calico.EnableIPV4POOL_NAT_OUTGOING(),
			"IPV6POOLNATOUTGOING":     g.KubeConf.Cluster.Network.Calico.EnableIPV6POOL_NAT_OUTGOING(),
			"DefaultIPPOOL":           g.KubeConf.Cluster.Network.Calico.EnableDefaultIPPOOL(),
			"IPv6Support":             IPv6Support,
			"NodeCidrMaskSizeIPv6":    g.KubeConf.Cluster.Kubernetes.NodeCidrMaskSizeIPv6,
			"TyphaReplicas":           g.KubeConf.Cluster.Network.Calico.Typha.Replicas,
			"TyphaNodeSelector":       g.KubeConf.Cluster.Network.Calico.Typha.NodeSelector,
			"ControllerReplicas":      g.KubeConf.Cluster.Network.Calico.Controller.Replicas,
			"ControllerNodeSelector":  g.KubeConf.Cluster.Network.Calico.Controller.NodeSelector,
		},
	}
	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	return nil
}
