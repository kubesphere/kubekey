package kubernetes

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
)

type NoClusterInfo struct {
	common.KubePrepare
}

func (n *NoClusterInfo) PreCheck(_ connector.Runtime) (bool, error) {
	if v, ok := n.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		if cluster.ClusterInfo == "" {
			return true, nil
		}
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
	return false, nil
}

type NodeInCluster struct {
	common.KubePrepare
	Not bool
}

func (n *NodeInCluster) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := n.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*KubernetesStatus)
		var versionOk bool
		if res, ok := cluster.NodesInfo[host.GetName()]; ok && res != "" {
			versionOk = true
		}
		_, ipOk := cluster.NodesInfo[host.GetInternalAddress()]
		if n.Not {
			return !(versionOk || ipOk), nil
		}
		return versionOk || ipOk, nil
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
}

type ClusterIsExist struct {
	common.KubePrepare
	Not bool
}

func (c *ClusterIsExist) PreCheck(_ connector.Runtime) (bool, error) {
	if exist, ok := c.PipelineCache.GetMustBool(common.ClusterExist); ok {
		if c.Not {
			return !exist, nil
		}
		return exist, nil
	} else {
		return false, errors.New("get kubernetes cluster status by pipeline cache failed")
	}
}

type ClusterNotEqualDesiredVersion struct {
	common.KubePrepare
}

func (c *ClusterNotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	clusterK8sVersion, ok := c.PipelineCache.GetMustString(common.K8sVersion)
	if !ok {
		return false, errors.New("get cluster Kubernetes version failed by pipeline cache")
	}

	if c.KubeConf.Cluster.Kubernetes.Version == clusterK8sVersion {
		return false, nil
	}
	return true, nil
}

type NotEqualDesiredVersion struct {
	common.KubePrepare
}

func (n *NotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()

	nodeK8sVersion, ok := host.GetCache().GetMustString(common.NodeK8sVersion)
	if !ok {
		return false, errors.New("get node Kubernetes version failed by host cache")
	}

	if n.KubeConf.Cluster.Kubernetes.Version == nodeK8sVersion {
		return false, nil
	}
	return true, nil
}
