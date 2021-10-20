package dns

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"strings"
)

type CoreDNSExist struct {
	common.KubePrepare
	Not bool
}

func (c *CoreDNSExist) PreCheck(runtime connector.Runtime) (bool, error) {
	_, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get svc -n kube-system coredns", false)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return c.Not, nil
		}
		return false, err
	}
	return !c.Not, nil
}

type EnableNodeLocalDNS struct {
	common.KubePrepare
}

func (e *EnableNodeLocalDNS) PreCheck(runtime connector.Runtime) (bool, error) {
	if e.KubeConf.Cluster.Kubernetes.EnableNodelocaldns() {
		return true, nil
	}
	return false, nil
}

type NodeLocalDNSConfigMapNotExist struct {
	common.KubePrepare
}

func (n *NodeLocalDNSConfigMapNotExist) PreCheck(runtime connector.Runtime) (bool, error) {
	if _, err := runtime.GetRunner().SudoCmd("/usr/local/bin/kubectl get cm -n kube-system nodelocaldns", false); err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return true, nil
		}
		return false, err
	}
	return false, nil
}
