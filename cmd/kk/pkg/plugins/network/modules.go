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
	"path/filepath"

	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/network/templates"
)

type DeployNetworkPluginModule struct {
	common.KubeModule
}

func (d *DeployNetworkPluginModule) Init() {
	d.Name = "DeployNetworkPluginModule"
	d.Desc = "Deploy cluster network plugin"

	switch d.KubeConf.Cluster.Network.Plugin {
	case common.Calico:
		d.Tasks = deployCalico(d)
	case common.Flannel:
		d.Tasks = deployFlannel(d)
	case common.Cilium:
		d.Tasks = deployCilium(d)
	case common.Kubeovn:
		d.Tasks = deployKubeOVN(d)
	case common.Hybridnet:
		d.Tasks = deployHybridnet(d)
	default:
		return
	}
	if d.KubeConf.Cluster.Network.EnableMultusCNI() {
		d.Tasks = append(d.Tasks, deployMultus(d)...)
	}
}

func deployMultus(d *DeployNetworkPluginModule) []task.Interface {
	generateMultus := &task.RemoteTask{
		Name:  "GenerateMultus",
		Desc:  "Generate multus cni",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&OldK8sVersion{Not: true},
		},
		Action: &action.Template{
			Template: templates.Multus,
			Dst:      filepath.Join(common.KubeConfigDir, templates.Multus.Name()),
			Data: util.Data{
				"MultusImage": images.GetImage(d.Runtime, d.KubeConf, "multus").ImageName(),
			},
		},
		Parallel: true,
	}
	deploy := &task.RemoteTask{
		Name:     "DeployMultus",
		Desc:     "Deploy multus",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkMultusPlugin),
		Parallel: true,
		Retry:    5,
	}
	return []task.Interface{
		generateMultus,
		deploy,
	}
}

func deployCalico(d *DeployNetworkPluginModule) []task.Interface {
	generateCalicoOld := &task.RemoteTask{
		Name:  "GenerateCalico",
		Desc:  "Generate calico",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(OldK8sVersion),
		},
		Action: &action.Template{
			Template: templates.CalicoOld,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CalicoOld.Name()),
			Data: util.Data{
				"KubePodsCIDR":           d.KubeConf.Cluster.Network.KubePodsCIDR,
				"CalicoCniImage":         images.GetImage(d.Runtime, d.KubeConf, "calico-cni").ImageName(),
				"CalicoNodeImage":        images.GetImage(d.Runtime, d.KubeConf, "calico-node").ImageName(),
				"CalicoFlexvolImage":     images.GetImage(d.Runtime, d.KubeConf, "calico-flexvol").ImageName(),
				"CalicoControllersImage": images.GetImage(d.Runtime, d.KubeConf, "calico-kube-controllers").ImageName(),
				"TyphaEnabled":           len(d.Runtime.GetHostsByRole(common.K8s)) > 50 || d.KubeConf.Cluster.Network.Calico.Typha(),
				"VethMTU":                d.KubeConf.Cluster.Network.Calico.VethMTU,
				"NodeCidrMaskSize":       d.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"IPIPMode":               d.KubeConf.Cluster.Network.Calico.IPIPMode,
				"VXLANMode":              d.KubeConf.Cluster.Network.Calico.VXLANMode,
			},
		},
		Parallel: true,
	}

	generateCalicoNew := &task.RemoteTask{
		Name:  "GenerateCalico",
		Desc:  "Generate calico",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&OldK8sVersion{Not: true},
		},
		Action: &action.Template{
			Template: templates.CalicoNew,
			Dst:      filepath.Join(common.KubeConfigDir, templates.CalicoNew.Name()),
			Data: util.Data{
				"KubePodsCIDR":            d.KubeConf.Cluster.Network.KubePodsCIDR,
				"CalicoCniImage":          images.GetImage(d.Runtime, d.KubeConf, "calico-cni").ImageName(),
				"CalicoNodeImage":         images.GetImage(d.Runtime, d.KubeConf, "calico-node").ImageName(),
				"CalicoFlexvolImage":      images.GetImage(d.Runtime, d.KubeConf, "calico-flexvol").ImageName(),
				"CalicoControllersImage":  images.GetImage(d.Runtime, d.KubeConf, "calico-kube-controllers").ImageName(),
				"CalicoTyphaImage":        images.GetImage(d.Runtime, d.KubeConf, "calico-typha").ImageName(),
				"TyphaEnabled":            len(d.Runtime.GetHostsByRole(common.K8s)) > 50 || d.KubeConf.Cluster.Network.Calico.Typha(),
				"VethMTU":                 d.KubeConf.Cluster.Network.Calico.VethMTU,
				"NodeCidrMaskSize":        d.KubeConf.Cluster.Kubernetes.NodeCidrMaskSize,
				"IPIPMode":                d.KubeConf.Cluster.Network.Calico.IPIPMode,
				"VXLANMode":               d.KubeConf.Cluster.Network.Calico.VXLANMode,
				"ConatinerManagerIsIsula": d.KubeConf.Cluster.Kubernetes.ContainerManager == "isula",
				"IPV4POOLNATOUTGOING":     d.KubeConf.Cluster.Network.Calico.EnableIPV4POOL_NAT_OUTGOING(),
				"DefaultIPPOOL":           d.KubeConf.Cluster.Network.Calico.EnableDefaultIPPOOL(),
			},
		},
		Parallel: true,
	}

	deploy := &task.RemoteTask{
		Name:     "DeployCalico",
		Desc:     "Deploy calico",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkPlugin),
		Parallel: true,
		Retry:    5,
	}

	if K8sVersionAtLeast(d.KubeConf.Cluster.Kubernetes.Version, "v1.16.0") {
		return []task.Interface{
			generateCalicoNew,
			deploy,
		}
	} else {
		return []task.Interface{
			generateCalicoOld,
			deploy,
		}
	}
}

func deployFlannel(d *DeployNetworkPluginModule) []task.Interface {
	generateFlannelPSP := &task.RemoteTask{
		Name:    "GenerateFlannel",
		Desc:    "Generate flannel",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action: &action.Template{
			Template: templates.FlannelPSP,
			Dst:      filepath.Join(common.KubeConfigDir, templates.FlannelPSP.Name()),
			Data: util.Data{
				"KubePodsCIDR":       d.KubeConf.Cluster.Network.KubePodsCIDR,
				"FlannelImage":       images.GetImage(d.Runtime, d.KubeConf, "flannel").ImageName(),
				"FlannelPluginImage": images.GetImage(d.Runtime, d.KubeConf, "flannel-cni-plugin").ImageName(),
				"BackendMode":        d.KubeConf.Cluster.Network.Flannel.BackendMode,
			},
		},
		Parallel: true,
	}
	generateFlannelPS := &task.RemoteTask{
		Name:    "GenerateFlannel",
		Desc:    "Generate flannel",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action: &action.Template{
			Template: templates.FlannelPS,
			Dst:      filepath.Join(common.KubeConfigDir, templates.FlannelPS.Name()),
			Data: util.Data{
				"KubePodsCIDR":       d.KubeConf.Cluster.Network.KubePodsCIDR,
				"FlannelImage":       images.GetImage(d.Runtime, d.KubeConf, "flannel").ImageName(),
				"FlannelPluginImage": images.GetImage(d.Runtime, d.KubeConf, "flannel-cni-plugin").ImageName(),
				"BackendMode":        d.KubeConf.Cluster.Network.Flannel.BackendMode,
			},
		},
		Parallel: true,
	}

	deploy := &task.RemoteTask{
		Name:     "DeployFlannel",
		Desc:     "Deploy flannel",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployNetworkPlugin),
		Parallel: true,
		Retry:    5,
	}

	if K8sVersionAtLeast(d.KubeConf.Cluster.Kubernetes.Version, "v1.25.0") {
		return []task.Interface{
			generateFlannelPS,
			deploy,
		}
	} else {
		return []task.Interface{
			generateFlannelPSP,
			deploy,
		}
	}
}

func deployCilium(d *DeployNetworkPluginModule) []task.Interface {

	releaseCiliumChart := &task.LocalTask{
		Name:   "GenerateCiliumChart",
		Desc:   "Generate cilium chart",
		Action: new(ReleaseCiliumChart),
	}

	syncCiliumChart := &task.RemoteTask{
		Name:     "SyncCiliumChart",
		Desc:     "Synchronize cilium chart",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(SyncCiliumChart),
		Parallel: true,
		Retry:    2,
	}

	deploy := &task.RemoteTask{
		Name:     "DeployCilium",
		Desc:     "Deploy cilium",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployCilium),
		Parallel: true,
		Retry:    5,
	}

	return []task.Interface{
		releaseCiliumChart,
		syncCiliumChart,
		deploy,
	}
}

func deployKubeOVN(d *DeployNetworkPluginModule) []task.Interface {
	label := &task.RemoteTask{
		Name:     "LabelNode",
		Desc:     "Label node",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(LabelNode),
		Parallel: true,
	}

	ssl := &task.RemoteTask{
		Name:  "GenerateSSl",
		Desc:  "Generate ssl",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(EnableSSL),
		},
		Action:   new(GenerateSSL),
		Parallel: true,
	}

	generateKubeOVN := &task.RemoteTask{
		Name:     "GenerateKubeOVN",
		Desc:     "Generate kube-ovn",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GenerateKubeOVN),
		Parallel: true,
	}

	deploy := &task.RemoteTask{
		Name:     "DeployKubeOVN",
		Desc:     "Deploy kube-ovn",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployKubeovnPlugin),
		Parallel: true,
		Retry:    5,
	}

	kubectlKo := &task.RemoteTask{
		Name:  "GenerateKubectlKo",
		Desc:  "Generate kubectl-ko",
		Hosts: d.Runtime.GetHostsByRole(common.Master),
		Action: &action.Template{
			Template: templates.KubectlKo,
			Dst:      filepath.Join(common.BinDir, templates.KubectlKo.Name()),
		},
		Parallel: true,
	}

	chmod := &task.RemoteTask{
		Name:     "ChmodKubectlKo",
		Desc:     "Chmod kubectl-ko",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Action:   new(ChmodKubectlKo),
		Parallel: true,
	}

	return []task.Interface{
		label,
		ssl,
		generateKubeOVN,
		deploy,
		kubectlKo,
		chmod,
	}
}

func deployHybridnet(d *DeployNetworkPluginModule) []task.Interface {

	releaseHybridnetChart := &task.LocalTask{
		Name:   "GenerateHybridnetChart",
		Desc:   "Generate hybridnet chart",
		Action: new(ReleaseHybridnetChart),
	}

	syncHybridnetChart := &task.RemoteTask{
		Name:     "SyncHybridnetChart",
		Desc:     "Synchronize hybridnet chart",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(SyncHybridnetChart),
		Parallel: true,
		Retry:    2,
	}

	deploy := &task.RemoteTask{
		Name:     "DeployHybridnet",
		Desc:     "Deploy hybridnet",
		Hosts:    d.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(DeployHybridnet),
		Parallel: true,
		Retry:    5,
	}

	return []task.Interface{
		releaseHybridnetChart,
		syncHybridnetChart,
		deploy,
	}
}

func K8sVersionAtLeast(version string, compare string) bool {
	cmp, err := versionutil.MustParseSemantic(version).Compare(compare)
	if err != nil {
		logger.Log.Fatal("unknown kubernetes version")
	}
	// old version
	if cmp == -1 {
		return false
	}
	// new version
	return true
}
