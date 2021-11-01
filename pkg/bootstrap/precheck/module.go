package precheck

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
)

type NodePreCheckModule struct {
	common.KubeModule
	Skip bool
}

func (n *NodePreCheckModule) IsSkip() bool {
	return n.Skip
}

func (n *NodePreCheckModule) Init() {
	n.Name = "NodePreCheckModule"
	n.Desc = "Do pre-check on cluster nodes"

	preCheck := &task.RemoteTask{
		Name:  "NodePreCheck",
		Desc:  "A pre-check on nodes",
		Hosts: n.Runtime.GetAllHosts(),
		//Prepare: &prepare.FastPrepare{
		//	Inject: func(runtime connector.Runtime) (bool, error) {
		//		if len(n.Runtime.GetHostsByRole(common.ETCD))%2 == 0 {
		//			logger.Log.Error("The number of etcd is even. Please configure it to be odd.")
		//			return false, errors.New("the number of etcd is even")
		//		}
		//		return true, nil
		//	}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	n.Tasks = []task.Interface{
		preCheck,
	}
}

type ClusterPreCheckModule struct {
	common.KubeModule
}

func (c *ClusterPreCheckModule) Init() {
	c.Name = "ClusterPreCheckModule"
	c.Desc = "Do pre-check on cluster"

	getKubeConfig := &task.RemoteTask{
		Name:     "GetKubeConfig",
		Desc:     "Get KubeConfig file",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GetKubeConfig),
		Parallel: true,
	}

	getAllNodesK8sVersion := &task.RemoteTask{
		Name:     "GetAllNodesK8sVersion",
		Desc:     "Get all nodes Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GetAllNodesK8sVersion),
		Parallel: true,
	}

	calculateMinK8sVersion := &task.RemoteTask{
		Name:     "CalculateMinK8sVersion",
		Desc:     "Calculate min Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CalculateMinK8sVersion),
		Parallel: true,
	}

	checkDesiredK8sVersion := &task.RemoteTask{
		Name:     "CheckDesiredK8sVersion",
		Desc:     "Check desired Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CheckDesiredK8sVersion),
		Parallel: true,
	}

	ksVersionCheck := &task.RemoteTask{
		Name:     "KsVersionCheck",
		Desc:     "Check KubeSphere version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(KsVersionCheck),
		Parallel: true,
	}

	dependencyCheck := &task.RemoteTask{
		Name:  "DependencyCheck",
		Desc:  "Check dependency matrix for KubeSphere and Kubernetes",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(KubeSphereExist),
		},
		Action:   new(DependencyCheck),
		Parallel: true,
	}

	getKubernetesNodesStatus := &task.RemoteTask{
		Name:     "GetKubernetesNodesStatus",
		Desc:     "Get kubernetes nodes status",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GetKubernetesNodesStatus),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		getKubeConfig,
		getAllNodesK8sVersion,
		calculateMinK8sVersion,
		checkDesiredK8sVersion,
		ksVersionCheck,
		dependencyCheck,
		getKubernetesNodesStatus,
	}
}
