package network

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type OldK8sVersion struct {
	common.KubePrepare
	Not bool
}

func (o *OldK8sVersion) PreCheck(_ connector.Runtime) (bool, error) {
	cmp, err := versionutil.MustParseSemantic(o.KubeConf.Cluster.Kubernetes.Version).Compare("v1.16.0")
	if err != nil {
		return false, err
	}
	// old version
	if cmp == -1 {
		return !o.Not, nil
	} else {
		// new version
		return o.Not, nil
	}
}
