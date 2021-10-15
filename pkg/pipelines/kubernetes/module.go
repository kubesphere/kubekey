package kubernetes

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/logger"
	"github.com/kubesphere/kubekey/pkg/core/modules"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/files"
	"github.com/kubesphere/kubekey/pkg/pipelines/binaries"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/images"
	"github.com/kubesphere/kubekey/pkg/pipelines/kubernetes/templates"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"os"
	"path/filepath"
	"sort"
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
		Desc:  "get kubernetes cluster status",
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
		Desc:     "synchronize kubernetes binaries",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	syncKubelet := &modules.RemoteTask{
		Name:     "SyncKubelet",
		Desc:     "synchronize kubelet",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(SyncKubelet),
		Parallel: true,
		Retry:    2,
	}

	generateKubeletService := &modules.RemoteTask{
		Name:    "GenerateKubeletService",
		Desc:    "generate kubelet service",
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
		Desc:     "enable kubelet service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Prepare:  &NodeInCluster{Not: true},
		Action:   new(EnableKubelet),
		Parallel: true,
		Retry:    5,
	}

	generateKubeletEnv := &modules.RemoteTask{
		Name:     "GenerateKubeletEnv",
		Desc:     "generate kubelet env",
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
		Desc:  "generate kubeadm config",
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
		Desc:  "init cluster using kubeadm",
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
		Desc:  "copy admin.conf to ~/.kube/config",
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
		Desc:  "remove master taint",
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
		Desc:  "add worker label",
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
		Desc:  "generate kubeadm config",
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
		Desc:  "copy admin.conf to ~/.kube/config",
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
		Desc:  "remove master taint",
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
		Desc:  "add worker label to master",
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
		Desc:  "synchronize kube config to worker",
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
		Desc:  "add worker label to worker",
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

type ProgressiveUpgradeModule struct {
	common.KubeModule
	Step UpgradeStep
}

func (p *ProgressiveUpgradeModule) Init() {
	p.Name = fmt.Sprintf("ProgressiveUpgradeModule %d/%d", p.Step, len(UpgradeStepList))

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
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(images.PullImage),
		Parallel: true,
	}

	syncBinary := &modules.RemoteTask{
		Name:     "SyncKubeBinary",
		Desc:     "Synchronize kubernetes binaries",
		Hosts:    p.Runtime.GetHostsByRole(common.K8s),
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(SyncKubeBinary),
		Parallel: true,
		Retry:    2,
	}

	upgradeKubeMaster := &modules.RemoteTask{
		Name:     "UpgradeClusterOnMaster",
		Desc:     "Upgrade cluster on master",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualDesiredVersion),
		Action:   &UpgradeKubeMaster{ModuleName: p.Name},
		Parallel: false,
	}

	cluster := NewKubernetesStatus()
	p.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &modules.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "get kubernetes cluster status",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(NotEqualDesiredVersion),
		Action:   new(GetClusterStatus),
		Parallel: false,
	}

	upgradeKubeWorker := &modules.RemoteTask{
		Name:  "UpgradeClusterOnWorker",
		Desc:  "Upgrade cluster on worker",
		Hosts: p.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(NotEqualDesiredVersion),
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
			new(NotEqualDesiredVersion),
		},
		Action:   &ReconfigureDNS{ModuleName: p.Name},
		Parallel: false,
	}

	p.Tasks = []modules.Task{
		download,
		pull,
		syncBinary,
		upgradeKubeMaster,
		clusterStatus,
		upgradeKubeWorker,
		reconfigureDNS,
	}
}

func (p *ProgressiveUpgradeModule) Run() error {
	currentVersion, ok := p.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return errors.New("get current Kubernetes version failed by pipeline cache")
	}
	desiredVersion := p.KubeConf.Cluster.Kubernetes.Version
	originalDesired := desiredVersion

	if cmp, err := versionutil.MustParseSemantic(currentVersion).Compare(desiredVersion); err != nil {
		return err
	} else if cmp == 1 {
		logger.Log.Messagef(
			common.LocalHost,
			"The current version (%s) is greater than the target version (%s)",
			currentVersion, desiredVersion)
		os.Exit(0)
	}

	if p.Step == ToV121 {
		v122 := versionutil.MustParseSemantic("v1.22.0")
		atLeast := versionutil.MustParseSemantic(desiredVersion).AtLeast(v122)
		cmp, err := versionutil.MustParseSemantic(currentVersion).Compare("v1.22.0")
		if err != nil {
			return err
		}
		if atLeast && cmp <= 0 {
			desiredVersion = "v1.21.5"
		}
	}

	end := false
	for !end {
		var nextVersionStr string
		if currentVersion != desiredVersion {
			nextVersionStr = calculateNextStr(currentVersion, desiredVersion)
			//u.PipelineCache.Set(common.DesiredK8sVersion, nextVersionStr)
			p.KubeConf.Cluster.Kubernetes.Version = nextVersionStr

			for i := range p.Tasks {
				task := p.Tasks[i]
				task.Init(p.Name, p.Runtime.(connector.Runtime), p.ModuleCache, p.PipelineCache)
				if res := task.Execute(); res.IsFailed() {
					return errors.Wrapf(res.CombineErr(), "Module[%s] exec failed", p.Name)
				}
			}

			currentVersion = nextVersionStr
			p.PipelineCache.Set(common.K8sVersion, nextVersionStr)
		} else {
			//u.PipelineCache.Set(common.DesiredK8sVersion, desiredVersion)
			p.KubeConf.Cluster.Kubernetes.Version = originalDesired
			end = true
		}
	}

	return nil
}

func calculateNextStr(currentVersion, desiredVersion string) string {
	current := versionutil.MustParseSemantic(currentVersion)
	target := versionutil.MustParseSemantic(desiredVersion)
	var nextVersionMinor uint
	if target.Minor() == current.Minor() {
		nextVersionMinor = current.Minor()
	} else {
		nextVersionMinor = current.Minor() + 1
	}

	if nextVersionMinor == target.Minor() {
		return desiredVersion
	} else {
		nextVersionPatchList := make([]int, 0)
		for supportVersionStr := range files.FileSha256["kubeadm"]["amd64"] {
			supportVersion := versionutil.MustParseSemantic(supportVersionStr)
			if supportVersion.Minor() == nextVersionMinor {
				nextVersionPatchList = append(nextVersionPatchList, int(supportVersion.Patch()))
			}
		}
		sort.Ints(nextVersionPatchList)

		nextVersion := current.WithMinor(nextVersionMinor)
		nextVersion = nextVersion.WithPatch(uint(nextVersionPatchList[len(nextVersionPatchList)-1]))

		return fmt.Sprintf("v%s", nextVersion.String())
	}
}
