package kubesphere

import (
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type VersionBelowV3 struct {
	common.KubePrepare
}

func (v *VersionBelowV3) PreCheck(runtime connector.Runtime) (bool, error) {
	versionStr, ok := v.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return false, errors.New("get current kubesphere version failed by pipeline cache")
	}
	version := versionutil.MustParseSemantic(versionStr)
	v300 := versionutil.MustParseSemantic("v3.0.0")
	if v.KubeConf.Cluster.KubeSphere.Enabled && v.KubeConf.Cluster.KubeSphere.Version == "v3.0.0" && version.LessThan(v300) {
		return true, nil
	}
	return false, nil
}

type NotEqualDesiredVersion struct {
	common.KubePrepare
}

func (n *NotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	ksVersion, ok := n.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		ksVersion = ""
	}

	if n.KubeConf.Cluster.KubeSphere.Version == ksVersion {
		return false, nil
	}
	return true, nil
}
