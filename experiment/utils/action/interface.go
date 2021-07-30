package action

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Action interface {
	Execute(node *kubekeyapiv1alpha1.HostCfg, vars pipeline.Vars) (result *pipeline.Result)
	Init(mgr *config.Manager)
}
