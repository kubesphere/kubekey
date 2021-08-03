package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Action interface {
	Execute(vars pipeline.Vars) (result *pipeline.Result)
	Init(mgr *config.Manager)
}
