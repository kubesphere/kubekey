package precheck

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/pkg/errors"
)

type KubeSphereExist struct {
	common.KubePrepare
}

func (k *KubeSphereExist) PreCheck(runtime connector.Runtime) (bool, error) {
	currentKsVersion, ok := k.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return false, errors.New("get current KubeSphere version failed by pipeline cache")
	}
	if currentKsVersion != "" {
		return true, nil
	}
	return false, nil
}
