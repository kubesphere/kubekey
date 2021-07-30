package action

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Template struct {
	mgr  *config.Manager
	Dst  string
	Data map[string]interface{}
}

func (t *Template) Init(mgr *config.Manager) {
	t.mgr = mgr
}

func (t *Template) Execute(node *kubekeyapiv1alpha1.HostCfg, vars pipeline.Vars) *pipeline.Result {
	fmt.Println(t.Data)
	return &pipeline.Result{}
}
