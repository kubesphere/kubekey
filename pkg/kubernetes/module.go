package kubernetes

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/binaries"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubernetes/templates"
	"github.com/pkg/errors"
	"path/filepath"
)

type KubernetesStatusModule struct {
	common.KubeModule
}

func (k *KubernetesStatusModule) Init() {
	k.Name = "KubernetesStatusModule"

	cluster := NewKubernetesStatus()
	k.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &modules.RemoteTask{
		Name:  "GetClusterStatus",
		Desc:  "Get kubernetes cluster status",
		Hosts: k.Runtime.GetHostsByRole(common.Master),
		//Prepare:  new(NoClusterInfo),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	k.Tasks = []modules.Task{
		clusterStatus,
	}
}

type InstallKubeBinariesModule struct {
	common.KubeModule
}

func (i *InstallKubeBinariesModule) Init() {
	i.Name = "InstallKubeBinariesModule"

	syncBinary := &modules.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	syncKubelet := &modules.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "Synchronize kubelet",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &modules.RemoteTask{
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

	enableKubelet := &modules.RemoteTask{
		Name:     "EnableKubelet",
		Desc:     "Enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &modules.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "Generate kubelet env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(GenerateKubeletEnv),
		Parallel: true,
		Retry:    2,
	}

	i.Tasks = []modules.Task{
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

	generateKubeadmConfig := &modules.RemoteTask{
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

	kubeadmInit := &modules.RemoteTask{
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

	copyKubeConfig := &modules.RemoteTask{
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

	removeMasterTaint := &modules.RemoteTask{
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

	addWorkerLabel := &modules.RemoteTask{
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

	i.Tasks = []modules.Task{
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

	j.PipelineCache.Set(common.ClusterExist, true)

	generateKubeadmConfig := &modules.RemoteTask{
		Name:  "GenerateKubeadmConfig",
		Desc:  "Generate kubeadm config",
		Hosts: j.Runtime.GetHostsByRole(common.K8s),
		Prepare: &prepare.PrepareCollection{
			&NodeInCluster{Not: true},
		},
		Action:   &GenerateKubeadmConfig{IsInitConfiguration: false},
		Parallel: true,
	}

	joinMasterNode := &modules.RemoteTask{
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

	joinWorkerNode := &modules.RemoteTask{
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

	copyKubeConfig := &modules.RemoteTask{
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

	removeMasterTaint := &modules.RemoteTask{
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

	addWorkerLabelToMaster := &modules.RemoteTask{
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

	syncKubeConfig := &modules.RemoteTask{
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

	addWorkerLabelToWorker := &modules.RemoteTask{
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

	j.Tasks = []modules.Task{
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

	kubeadmReset := &modules.RemoteTask{
		Name:     "KubeadmReset",
		Desc:     "Reset the cluster using kubeadm",
		Hosts:    r.Runtime.GetHostsByRole(common.K8s),
		Action:   new(KubeadmReset),
		Parallel: true,
	}

	r.Tasks = []modules.Task{
		kubeadmReset,
	}
}

type CompareConfigAndClusterInfoModule struct {
	common.KubeModule
}

func (c *CompareConfigAndClusterInfoModule) Init() {
	c.Name = "CompareConfigAndClusterInfoModule"

	check := &modules.RemoteTask{
		Name:    "FindDifferences",
		Desc:    "Find the differences between config and cluster node info",
		Hosts:   c.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(FindDifferences),
	}

	c.Tasks = []modules.Task{
		check,
	}
}

type DeleteKubeNodeModule struct {
	common.KubeModule
}

func (r *DeleteKubeNodeModule) Init() {
	r.Name = "DeleteKubeNodeModule"

	drain := &modules.RemoteTask{
		Name:    "DrainNode",
		Desc:    "Node safely evict all pods",
		Hosts:   r.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(DrainNode),
		Retry:   5,
	}

	deleteNode := &modules.RemoteTask{
		Name:    "DeleteNode",
		Desc:    "Delete the node using kubectl",
		Hosts:   r.Runtime.GetHostsByRole(common.Master),
		Prepare: new(common.OnlyFirstMaster),
		Action:  new(KubectlDeleteNode),
		Retry:   5,
	}

	r.Tasks = []modules.Task{
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

	plan := &modules.LocalTask{
		Name:   "SetUpgradePlan",
		Desc:   "Set upgrade plan",
		Action: &SetUpgradePlan{Step: s.Step},
	}

	s.Tasks = []modules.Task{
		plan,
	}
}

type ProgressiveUpgradeModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (p *ProgressiveUpgradeModule) Init() {
	p.Name = fmt.Sprintf("ProgressiveUpgradeModule %d/%d", p.Step, len(UpgradeStepList))

	nextVersion := &modules.LocalTask{
		Name:    "CalculateNextVersion",
		Desc:    "Calculate next upgrade version",
		Prepare: new(ClusterNotEqualDesiredVersion),
		Action:  new(CalculateNextVersion),
	}

	download := &modules.LocalTask{
		Name:    "DownloadBinaries",
		Desc:    "Download installation binaries",
		Prepare: new(ClusterNotEqualDesiredVersion),
		Action:  new(binaries.Download),
	}

	pull := &modules.RemoteTask{
		Name:     "PullImages",
		Desc:     "Start to pull images on all nodes",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(ClusterNotEqualDesiredVersion),
		Action:   new(images.PullImage),
		Parallel: true,
	}

	syncBinary := &modules.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(ClusterNotEqualDesiredVersion),
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	upgradeKubeMaster := &modules.RemoteTask{
		Name:     "UpgradeClusterOnMaster",
		Desc:     "Upgrade cluster on master",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(ClusterNotEqualDesiredVersion),
		Action:   &UpgradeKubeMaster{ModuleName: p.Name},
		Parallel: false,
	}

	cluster := NewKubernetesStatus()
	p.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &modules.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get kubernetes cluster status",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(ClusterNotEqualDesiredVersion),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	upgradeKubeWorker := &modules.RemoteTask{
		Name:  "UpgradeClusterOnWorker",
		Desc:  "Upgrade cluster on worker",
		Hosts: p.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(ClusterNotEqualDesiredVersion),
			new(common.OnlyWorker),
		},
		Action:   &UpgradeKubeWorker{ModuleName: p.Name},
		Parallel: false,
	}

	reconfigureDNS := &modules.RemoteTask{
		Name:  "ReconfigureCoreDNS",
		Desc:  "Reconfigure CoreDNS",
		Hosts: p.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(ClusterNotEqualDesiredVersion),
		},
		Action:   &ReconfigureDNS{ModuleName: p.Name},
		Parallel: false,
	}

	currentVersion := &modules.LocalTask{
		Name:    "SetCurrentK8sVersion",
		Desc:    "Set current k8s version",
		Prepare: new(ClusterNotEqualDesiredVersion),
		Action:  new(SetCurrentK8sVersion),
	}

	p.Tasks = []modules.Task{
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
