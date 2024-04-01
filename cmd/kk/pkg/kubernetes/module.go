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

package kubernetes

import (
	"fmt"
	"path/filepath"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/dns"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/binaries"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/prepare"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/task"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/images"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/kubernetes/templates"
	dnsTemplates "github.com/kubesphere/kubekey/v3/cmd/kk/pkg/plugins/dns/templates"
)

type StatusModule struct {
	common.KubeModule
}

func (k *StatusModule) Init() {
	k.Name = "KubernetesStatusModule"
	k.Desc = "Get kubernetes cluster status"

	cluster := NewKubernetesStatus()
	k.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:  "GetClusterStatus",
		Desc:  "Get kubernetes cluster status",
		Hosts: k.Runtime.GetHostsByRole(common.Master),
		//Prepare:  new(NoClusterInfo),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	k.Tasks = []task.Interface{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"
	i.Desc = "Install kubernetes cluster"

	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	chmodKubelet := &task.RemoteTask{
		Name:     "ChmodKubelet",
		Desc:     "Change kubelet mode",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(ChmodKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &task.RemoteTask{
		Name:    "GenerateKubeletService",
		Desc:    "Generate kubelet service",
		Hosts:   i.Runtime.GetHostsByRole(common.K8s),
		Prepare: &NodeInCluster{Not: true},
		Action: &action.Template{
			Template: templates.KubeletService,
			Dst:      filepath.Join("/etc/systemd/system/", templates.KubeletService.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	enableKubelet := &task.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "Enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &task.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "Generate kubelet env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(GenerateKubeletEnv),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []task.Interface{
		syncBinary,
		chmodKubelet,
		generateKubeletService,
		enableKubelet,
		generateKubeletEnv,
	}
}

type InitKubernetesModule struct {
	common.KubeModule
}

func (i *InitKubernetesModule) Init() {
	i.Name = "InitKubernetesModule"
	i.Desc = "Init kubernetes cluster"

	generateKubeadmConfig := &task.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action: &GenerateKubeadmConfig{
			IsInitConfiguration:     true,
			WithSecurityEnhancement: i.KubeConf.Arg.SecurityEnhancement,
		},
		Parallel: true,
	}

	generateAuditPolicy := &task.RemoteTask{
		Name:  "GenerateAduitPolicy",
		Desc:  "Generate audit policy",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.EnableAudit),
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.AuditPolicy,
			Dst:      filepath.Join("/etc/kubernetes/audit", templates.AuditPolicy.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	generateAuditWebhook := &task.RemoteTask{
		Name:  "GenerateAduitWebhook",
		Desc:  "Generate audit webhook",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.EnableAudit),
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action: &action.Template{
			Template: templates.AuditWebhook,
			Dst:      filepath.Join("/etc/kubernetes/audit", templates.AuditWebhook.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	kubeadmInit := &task.RemoteTask{
		Name:  "KubeadmInit",
		Desc:  "Init cluster using kubeadm",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(KubeadmInit),
		Retry:    3,
		Parallel: true,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "CopyKubeConfig",
		Desc:  "Copy admin.conf to ~/.kube/config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: true,
	}

	removeMasterTaint := &task.RemoteTask{
		Name:  "RemoveMasterTaint",
		Desc:  "Remove master taint",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(RemoveMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	i.Tasks = []task.Interface{
		generateKubeadmConfig,
		generateAuditPolicy,
		generateAuditWebhook,
		kubeadmInit,
		copyKubeConfig,
		removeMasterTaint,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "JoinNodesModule"
	j.Desc = "Join kubernetes nodes"

	j.PipelineCache.Set(common.ClusterExist, true)

	generateKubeadmConfig := &task.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action: &GenerateKubeadmConfig{
			IsInitConfiguration:     false,
			WithSecurityEnhancement: j.KubeConf.Arg.SecurityEnhancement,
		},
		Parallel: true,
	}

	generateAuditPolicy := &task.RemoteTask{
		Name:  "GenerateAduitPolicy",
		Desc:  "Generate audit policy",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.EnableAudit),
			&NodeInCluster{Not: true},
		},
		Action: &action.Template{
			Template: templates.AuditPolicy,
			Dst:      filepath.Join("/etc/kubernetes/audit", templates.AuditPolicy.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	generateAuditWebhook := &task.RemoteTask{
		Name:  "GenerateAduitWebhook",
		Desc:  "Generate audit webhook",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.EnableAudit),
			&NodeInCluster{Not: true},
		},
		Action: &action.Template{
			Template: templates.AuditWebhook,
			Dst:      filepath.Join("/etc/kubernetes/audit", templates.AuditWebhook.Name()),
		},
		Parallel: true,
		Retry:    2,
	}

	joinMasterNode := &task.RemoteTask{
		Name:  "JoinControlPlaneNode",
		Desc:  "Join control-plane node",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(JoinNode),
		Parallel: true,
		Retry:    5,
	}

	joinWorkerNode := &task.RemoteTask{
		Name:  "JoinWorkerNode",
		Desc:  "Join worker node",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(JoinNode),
		Parallel: true,
		Retry:    5,
	}

	copyKubeConfig := &task.RemoteTask{
		Name:  "copyKubeConfig",
		Desc:  "Copy admin.conf to ~/.kube/config",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(CopyKubeConfigForControlPlane),
		Parallel: true,
		Retry:    2,
	}

	removeMasterTaint := &task.RemoteTask{
		Name:  "RemoveMasterTaint",
		Desc:  "Remove master taint",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(RemoveMasterTaint),
		Parallel: true,
		Retry:    5,
	}

	addWorkerLabelToNode := &task.RemoteTask{
		Name:    "addWorkerLabelToNode",
		Desc:    "Add worker label to all nodes",
		Hosts:   j.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(AddWorkerLabel),
		Retry:   3,
	}

	j.Tasks = []task.Interface{
		generateKubeadmConfig,
		generateAuditPolicy,
		generateAuditWebhook,
		joinMasterNode,
		joinWorkerNode,
		copyKubeConfig,
		removeMasterTaint,
		addWorkerLabelToNode,
	}
}

type ResetClusterModule struct {
	common.KubeModule
}

func (r *ResetClusterModule) Init() {
	r.Name = "ResetClusterModule"
	r.Desc = "Reset kubernetes cluster"

	kubeadmReset := &task.RemoteTask{
		Name:     "KubeadmReset",
		Desc:     "Reset the cluster using kubeadm",
		Hosts:    r.Runtime.GetHostsByRole(common.K8s),
		Action:   new(KubeadmReset),
		Parallel: true,
	}

	r.Tasks = []task.Interface{
		kubeadmReset,
	}
}

type CompareConfigAndClusterInfoModule struct {
	common.KubeModule
}

func (c *CompareConfigAndClusterInfoModule) Init() {
	c.Name = "CompareConfigAndClusterInfoModule"
	c.Desc = "Compare config and cluster nodes info"

	check := &task.RemoteTask{
		Name:    "FindNode",
		Desc:    "Find information about nodes that are expected to be deleted",
		Hosts:   c.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		//Action:  new(FindNode),
		Action: new(FilterFirstMaster),
	}

	c.Tasks = []task.Interface{
		check,
	}
}

type RestartKubeletModule struct {
	common.KubeModule
}

func (r *RestartKubeletModule) init() {
	r.Name = "RestartKubeletModule"
	r.Desc = "restart node kubelet service "
	restart := &task.RemoteTask{
		Name:   "RestartKubelet",
		Desc:   "Restart kubelet service",
		Hosts:  r.Runtime.GetHostsByRole(common.Master),
		Action: new(RestartKubelet),
		Retry:  5,
	}

	r.Tasks = []task.Interface{
		restart,
	}
}

type DeleteKubeNodeModule struct {
	common.KubeModule
}

func (d *DeleteKubeNodeModule) Init() {
	d.Name = "DeleteKubeNodeModule"
	d.Desc = "Delete kubernetes node"

	drain := &task.RemoteTask{
		Name:    "DrainNode",
		Desc:    "Node safely evict all pods",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(DrainNode),
		Retry:   2,
	}

	deleteNode := &task.RemoteTask{
		Name:    "DeleteNode",
		Desc:    "Delete the node using kubectl",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(KubectlDeleteNode),
		Retry:   5,
	}

	d.Tasks = []task.Interface{
		drain,
		deleteNode,
	}
}

type SetUpgradePlanModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (s *SetUpgradePlanModule) Init() {
	s.Name = fmt.Sprintf("SetUpgradePlanModule %d/%d", s.Step, len(UpgradeStepList))
	s.Desc = "Set upgrade plan"

	plan := &task.LocalTask{
		Name:   "SetUpgradePlan",
		Desc:   "Set upgrade plan",
		Action: &SetUpgradePlan{Step: s.Step},
	}

	generateKubeadmConfigInit := &task.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: s.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action: &GenerateKubeadmConfig{
			IsInitConfiguration:     true,
			WithSecurityEnhancement: s.KubeConf.Arg.SecurityEnhancement,
		},
		Parallel: true,
	}

	s.Tasks = []task.Interface{
		plan,
		generateKubeadmConfigInit,
	}
}

type ProgressiveUpgradeModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (p *ProgressiveUpgradeModule) Init() {
	p.Name = fmt.Sprintf("ProgressiveUpgradeModule %d/%d", p.Step, len(UpgradeStepList))
	p.Desc = fmt.Sprintf("Progressive upgrade %d/%d", p.Step, len(UpgradeStepList))

	nextVersion := &task.LocalTask{
		Name:    "CalculateNextVersion",
		Desc:    "Calculate next upgrade version",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(CalculateNextVersion),
	}

	// prepare
	download := &task.LocalTask{
		Name:    "DownloadBinaries",
		Desc:    "Download installation binaries",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(binaries.Download),
	}

	pull := &task.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(images.PullImage),
		Parallel: true,
	}

	// upgrade kubernetes
	syncBinary := &task.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	upgradeKubeMaster := &task.RemoteTask{
		Name:     "UpgradeClusterOnMaster",
		Desc:     "Upgrade cluster on master",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualPlanVersion),
		Action:   &UpgradeKubeMaster{ModuleName: p.Name},
		Parallel: false,
	}

	cluster := NewKubernetesStatus()
	p.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get kubernetes cluster status",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	generateCoreDNS := &task.RemoteTask{
		Name:  "GenerateCoreDNS",
		Desc:  "Generate coredns manifests",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualPlanVersion),
			new(common.OnlyFirstMaster),
		},
		Action:   new(dns.GenerateCorednsmanifests),
		Parallel: true,
	}

	applyCoredns := &task.RemoteTask{
		Name:  "DeployCoreDNS",
		Desc:  "Deploy coredns",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualPlanVersion),
			new(common.OnlyFirstMaster),
		},
		Action:   new(dns.DeployCoreDNS),
		Parallel: true,
	}

	generateNodeLocalDNS := &task.RemoteTask{
		Name:  "GenerateNodeLocalDNS",
		Desc:  "Generate nodelocaldns",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualPlanVersion),
			new(common.OnlyFirstMaster),
			new(dns.EnableNodeLocalDNS),
		},
		Action: &action.Template{
			Template: dnsTemplates.NodeLocalDNSService,
			Dst:      filepath.Join(common.KubeConfigDir, dnsTemplates.NodeLocalDNSService.Name()),
			Data: util.Data{
				"NodelocaldnsImage": images.GetImage(p.Runtime, p.KubeConf, "k8s-dns-node-cache").ImageName(),
				"DNSEtcHosts":       p.KubeConf.Cluster.DNS.DNSEtcHosts,
			},
		},
		Parallel: true,
	}

	applyNodeLocalDNS := &task.RemoteTask{
		Name:  "DeployNodeLocalDNS",
		Desc:  "Deploy nodelocaldns",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualPlanVersion),
			new(common.OnlyFirstMaster),
			new(dns.EnableNodeLocalDNS)},
		Action:   new(dns.DeployNodeLocalDNS),
		Parallel: true,
		Retry:    5,
	}

	upgradeKubeWorker := &task.RemoteTask{
		Name:  "UpgradeClusterOnWorker",
		Desc:  "Upgrade cluster on worker",
		Hosts: p.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualPlanVersion),
			new(common.OnlyWorker),
		},
		Action:   &UpgradeKubeWorker{ModuleName: p.Name},
		Parallel: false,
	}

	currentVersion := &task.LocalTask{
		Name:    "SetCurrentK8sVersion",
		Desc:    "Set current k8s version",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(SetCurrentK8sVersion),
	}

	p.Tasks = []task.Interface{
		nextVersion,
		download,
		pull,
		syncBinary,
		upgradeKubeMaster,
		clusterStatus,
		upgradeKubeWorker,
		generateCoreDNS,
		applyCoredns,
		generateNodeLocalDNS,
		applyNodeLocalDNS,
		currentVersion,
	}
}

func (p *ProgressiveUpgradeModule) Until() (*bool, error) {
	f := false
	t := true
	currentVersion, ok := p.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return &f, errors.New("get current Kubernetes version failed by pipeline cache")
	}
	planVersion, ok := p.PipelineCache.GetMustString(common.PlanK8sVersion)
	if !ok {
		return &f, errors.New("get upgrade plan Kubernetes version failed by pipeline cache")
	}

	if currentVersion != planVersion {
		return &f, nil
	} else {
		originalDesired, ok := p.PipelineCache.GetMustString(common.DesiredK8sVersion)
		if !ok {
			return &f, errors.New("get original desired Kubernetes version failed by pipeline cache")
		}
		p.KubeConf.Cluster.Kubernetes.Version = originalDesired
		return &t, nil
	}
}

type SaveKubeConfigModule struct {
	common.KubeModule
}

func (s *SaveKubeConfigModule) Init() {
	s.Name = "SaveKubeConfigModule"
	s.Desc = "Save kube config file as a configmap"

	save := &task.LocalTask{
		Name:   "SaveKubeConfig",
		Desc:   "Save kube config as a configmap",
		Action: new(SaveKubeConfig),
		Retry:  5,
	}

	s.Tasks = []task.Interface{
		save,
	}
}

type ConfigureKubernetesModule struct {
	common.KubeModule
}

func (c *ConfigureKubernetesModule) Init() {
	c.Name = "ConfigureKubernetesModule"
	c.Desc = "Configure kubernetes"

	configure := &task.RemoteTask{
		Name:    "ConfigureKubernetes",
		Desc:    "Configure kubernetes",
		Hosts:   c.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(ConfigureKubernetes),
		Retry:   3,
	}

	c.Tasks = []task.Interface{
		configure,
	}
}

type SecurityEnhancementModule struct {
	common.KubeModule
	Skip bool
}

func (s *SecurityEnhancementModule) IsSkip() bool {
	return s.Skip
}

func (s *SecurityEnhancementModule) Init() {
	s.Name = "SecurityEnhancementModule"
	s.Desc = "Security enhancement for the cluster"

	etcdSecurityEnhancement := &task.RemoteTask{
		Name:     "EtcdSecurityEnhancementTask",
		Desc:     "Security enhancement for etcd",
		Hosts:    s.Runtime.GetHostsByRole(common.ETCD),
		Action:   new(EtcdSecurityEnhancemenAction),
		Parallel: true,
	}

	masterSecurityEnhancement := &task.RemoteTask{
		Name:     "K8sSecurityEnhancementTask",
		Desc:     "Security enhancement for kubernetes",
		Hosts:    s.Runtime.GetHostsByRole(common.Master),
		Action:   new(MasterSecurityEnhancemenAction),
		Parallel: true,
	}

	nodesSecurityEnhancement := &task.RemoteTask{
		Name:     "K8sSecurityEnhancementTask",
		Desc:     "Security enhancement for kubernetes",
		Hosts:    s.Runtime.GetHostsByRole(common.Worker),
		Action:   new(NodesSecurityEnhancemenAction),
		Parallel: true,
	}

	s.Tasks = []task.Interface{
		etcdSecurityEnhancement,
		masterSecurityEnhancement,
		nodesSecurityEnhancement,
	}
}
