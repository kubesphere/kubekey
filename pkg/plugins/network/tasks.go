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
	"fmt"
	"github.com/kubesphere/kubekey/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/plugins/network/templates"
	"github.com/pkg/errors"
	"path/filepath"
)

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
		"/usr/local/bin/kubectl label no -lbeta.kubernetes.io/os=linux kubernetes.io/os=linux --overwrite",
		true); err != nil {
		return errors.Wrap(errors.WithStack(err), "override node label failed")
	}
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

type GenerateKubeOVNOld struct {
	common.KubeAction
}

func (g *GenerateKubeOVNOld) Execute(runtime connector.Runtime) error {
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
		Template: templates.KubeOVNOld,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOVNOld.Name()),
		Data: util.Data{
			"Address":             address,
			"Count":               count,
			"KubeovnImage":        images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"PodCIDR":             g.KubeConf.Cluster.Network.KubePodsCIDR,
			"SvcCIDR":             g.KubeConf.Cluster.Network.KubeServiceCIDR,
			"JoinCIDR":            g.KubeConf.Cluster.Network.Kubeovn.JoinCIDR,
			"PingExternalAddress": g.KubeConf.Cluster.Network.Kubeovn.PingerExternalAddress,
			"PingExternalDNS":     g.KubeConf.Cluster.Network.Kubeovn.PingerExternalDomain,
			"NetworkType":         g.KubeConf.Cluster.Network.Kubeovn.NetworkType,
			"VlanID":              g.KubeConf.Cluster.Network.Kubeovn.VlanID,
			"VlanInterfaceName":   g.KubeConf.Cluster.Network.Kubeovn.VlanInterfaceName,
			"Iface":               g.KubeConf.Cluster.Network.Kubeovn.Iface,
			"DpdkMode":            g.KubeConf.Cluster.Network.Kubeovn.DpdkMode,
			"DpdkVersion":         g.KubeConf.Cluster.Network.Kubeovn.DpdkVersion,
			"OvnVersion":          v1alpha2.DefaultKubeovnVersion,
			"EnableSSL":           g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"EnableMirror":        g.KubeConf.Cluster.Network.Kubeovn.EnableMirror,
			"HwOffload":           g.KubeConf.Cluster.Network.Kubeovn.HwOffload,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type GenerateKubeOVNNew struct {
	common.KubeAction
}

func (g *GenerateKubeOVNNew) Execute(runtime connector.Runtime) error {
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
		Template: templates.KubeOVNOld,
		Dst:      filepath.Join(common.KubeConfigDir, templates.KubeOVNOld.Name()),
		Data: util.Data{
			"Address":             address,
			"Count":               count,
			"KubeovnImage":        images.GetImage(runtime, g.KubeConf, "kubeovn").ImageName(),
			"PodCIDR":             g.KubeConf.Cluster.Network.KubePodsCIDR,
			"SvcCIDR":             g.KubeConf.Cluster.Network.KubeServiceCIDR,
			"JoinCIDR":            g.KubeConf.Cluster.Network.Kubeovn.JoinCIDR,
			"PingExternalAddress": g.KubeConf.Cluster.Network.Kubeovn.PingerExternalAddress,
			"PingExternalDNS":     g.KubeConf.Cluster.Network.Kubeovn.PingerExternalDomain,
			"NetworkType":         g.KubeConf.Cluster.Network.Kubeovn.NetworkType,
			"VlanID":              g.KubeConf.Cluster.Network.Kubeovn.VlanID,
			"VlanInterfaceName":   g.KubeConf.Cluster.Network.Kubeovn.VlanInterfaceName,
			"Iface":               g.KubeConf.Cluster.Network.Kubeovn.Iface,
			"DpdkMode":            g.KubeConf.Cluster.Network.Kubeovn.DpdkMode,
			"DpdkVersion":         g.KubeConf.Cluster.Network.Kubeovn.DpdkVersion,
			"OvnVersion":          v1alpha2.DefaultKubeovnVersion,
			"EnableSSL":           g.KubeConf.Cluster.Network.Kubeovn.EnableSSL,
			"EnableMirror":        g.KubeConf.Cluster.Network.Kubeovn.EnableMirror,
			"HwOffload":           g.KubeConf.Cluster.Network.Kubeovn.HwOffload,
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
