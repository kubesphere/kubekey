package kubernetes

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/module"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubernetes/templates"
	"github.com/pkg/errors"
	"path/filepath"
)

type StatusModule struct {
	common.KubeModule
}

func (k *StatusModule) Init() {
	k.Name = "KubernetesStatusModule"
	k.Desc = "Get kubernetes cluster status"

	cluster := NewKubernetesStatus()
	k.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &module.RemoteTask{
		Name:  "GetClusterStatus",
		Desc:  "Get kubernetes cluster status",
		Hosts: k.Runtime.GetHostsByRole(common.Master),
		//Prepare:  new(NoClusterInfo),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	k.Tasks = []module.Task{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"
	i.Desc = "Install kubernetes cluster"

	syncBinary := &module.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	syncKubelet := &module.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "Synchronize kubelet",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &module.RemoteTask{
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

	enableKubelet := &module.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "Enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &module.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "Generate kubelet env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(GenerateKubeletEnv),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []module.Task{
		syncBinary,
		syncKubelet,
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

	generateKubeadmConfig := &module.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
		},
		Action:   &GenerateKubeadmConfig{IsInitConfiguration: true},
		Parallel: true,
	}

	kubeadmInit := &module.RemoteTask{
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

	copyKubeConfig := &module.RemoteTask{
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

	removeMasterTaint := &module.RemoteTask{
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

	addWorkerLabel := &module.RemoteTask{
		Name:  "AddWorkerLabel",
		Desc:  "Add worker label",
		Hosts: i.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&ClusterIsExist{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	i.Tasks = []module.Task{
		generateKubeadmConfig,
		kubeadmInit,
		copyKubeConfig,
		removeMasterTaint,
		addWorkerLabel,
	}
}

type JoinNodesModule struct {
	common.KubeModule
}

func (j *JoinNodesModule) Init() {
	j.Name = "JoinNodesModule"
	j.Desc = "Join kubernetes nodes"

	j.PipelineCache.Set(common.ClusterExist, true)

	generateKubeadmConfig := &module.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   &GenerateKubeadmConfig{IsInitConfiguration: false},
		Parallel: true,
	}

	joinMasterNode := &module.RemoteTask{
		Name:  "JoinMasterNode",
		Desc:  "Join master node",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   new(JoinNode),
		Parallel: true,
		Retry:    5,
	}

	joinWorkerNode := &module.RemoteTask{
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

	copyKubeConfig := &module.RemoteTask{
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

	removeMasterTaint := &module.RemoteTask{
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

	addWorkerLabelToMaster := &module.RemoteTask{
		Name:  "AddWorkerLabelToMaster",
		Desc:  "Add worker label to master",
		Hosts: j.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.IsWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	syncKubeConfig := &module.RemoteTask{
		Name:  "SyncKubeConfig",
		Desc:  "Synchronize kube config to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(SyncKubeConfigToWorker),
		Parallel: true,
		Retry:    3,
	}

	addWorkerLabelToWorker := &module.RemoteTask{
		Name:  "AddWorkerLabelToWorker",
		Desc:  "Add worker label to worker",
		Hosts: j.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
			new(common.OnlyWorker),
		},
		Action:   new(AddWorkerLabel),
		Parallel: true,
		Retry:    5,
	}

	j.Tasks = []module.Task{
		generateKubeadmConfig,
		joinMasterNode,
		joinWorkerNode,
		copyKubeConfig,
		removeMasterTaint,
		addWorkerLabelToMaster,
		syncKubeConfig,
		addWorkerLabelToWorker,
	}
}

type ResetClusterModule struct {
	common.KubeModule
}

func (r *ResetClusterModule) Init() {
	r.Name = "ResetClusterModule"
	r.Desc = "Reset kubernetes cluster"

	kubeadmReset := &module.RemoteTask{
		Name:     "KubeadmReset",
		Desc:     "Reset the cluster using kubeadm",
		Hosts:    r.Runtime.GetHostsByRole(common.K8s),
		Action:   new(KubeadmReset),
		Parallel: true,
	}

	r.Tasks = []module.Task{
		kubeadmReset,
	}
}

type CompareConfigAndClusterInfoModule struct {
	common.KubeModule
}

func (c *CompareConfigAndClusterInfoModule) Init() {
	c.Name = "CompareConfigAndClusterInfoModule"
	c.Desc = "Compare config and cluster nodes info"

	check := &module.RemoteTask{
		Name:    "FindDifferences",
		Desc:    "Find the differences between config and cluster node info",
		Hosts:   c.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(FindDifferences),
	}

	c.Tasks = []module.Task{
		check,
	}
}

type DeleteKubeNodeModule struct {
	common.KubeModule
}

func (d *DeleteKubeNodeModule) Init() {
	d.Name = "DeleteKubeNodeModule"
	d.Desc = "Delete kubernetes node"

	drain := &module.RemoteTask{
		Name:    "DrainNode",
		Desc:    "Node safely evict all pods",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(DrainNode),
		Retry:   5,
	}

	deleteNode := &module.RemoteTask{
		Name:    "DeleteNode",
		Desc:    "Delete the node using kubectl",
		Hosts:   d.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(KubectlDeleteNode),
		Retry:   5,
	}

	d.Tasks = []module.Task{
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

	plan := &module.LocalTask{
		Name:   "SetUpgradePlan",
		Desc:   "Set upgrade plan",
		Action: &SetUpgradePlan{Step: s.Step},
	}

	s.Tasks = []module.Task{
		plan,
	}
}

type ProgressiveUpgradeModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (p *ProgressiveUpgradeModule) Init() {
	p.Name = fmt.Sprintf("ProgressiveUpgradeModule %d/%d", p.Step, len(UpgradeStepList))
	p.Desc = fmt.Sprintf("Progressive upgrade %d/%d", p.Step, len(UpgradeStepList))

	nextVersion := &module.LocalTask{
		Name:    "CalculateNextVersion",
		Desc:    "Calculate next upgrade version",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(CalculateNextVersion),
	}

	download := &module.LocalTask{
		Name:    "DownloadBinaries",
		Desc:    "Download installation binaries",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(binaries.Download),
	}

	pull := &module.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(images.PullImage),
		Parallel: true,
	}

	syncBinary := &module.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	upgradeKubeMaster := &module.RemoteTask{
		Name:     "UpgradeClusterOnMaster",
		Desc:     "Upgrade cluster on master",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualPlanVersion),
		Action:   &UpgradeKubeMaster{ModuleName: p.Name},
		Parallel: false,
	}

	cluster := NewKubernetesStatus()
	p.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &module.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get kubernetes cluster status",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualPlanVersion),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	upgradeKubeWorker := &module.RemoteTask{
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

	reconfigureDNS := &module.RemoteTask{
		Name:  "ReconfigureCoreDNS",
		Desc:  "Reconfigure CoreDNS",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualPlanVersion),
		},
		Action:   &ReconfigureDNS{ModuleName: p.Name},
		Parallel: false,
	}

	currentVersion := &module.LocalTask{
		Name:    "SetCurrentK8sVersion",
		Desc:    "Set current k8s version",
		Prepare: new(NotEqualPlanVersion),
		Action:  new(SetCurrentK8sVersion),
	}

	p.Tasks = []module.Task{
		nextVersion,
		download,
		pull,
		syncBinary,
		upgradeKubeMaster,
		clusterStatus,
		upgradeKubeWorker,
		reconfigureDNS,
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
