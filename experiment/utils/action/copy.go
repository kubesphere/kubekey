package action

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Copy struct {
	mgr    *config.Manager
	Src    string
	Dst    string
	Result pipeline.Result
}

func (c *Copy) Init(mgr *config.Manager) {
	c.mgr = mgr
}

func (c *Copy) Execute(node *kubekeyapiv1alpha1.HostCfg, vars pipeline.Vars) *pipeline.Result {
	fmt.Println(c.Dst, c.Src)
	return &pipeline.Result{}
}
