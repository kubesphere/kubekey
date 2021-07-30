package action

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type WebServer struct {
	ListenPort    int
	ListenAddress string // 127.0.0.1, 0.0.0.0, *
	Status        string // start, restart, stop
}

func (w *WebServer) Execute(node *kubekeyapiv1alpha1.HostCfg, vars pipeline.Vars) *pipeline.Result {
	fmt.Println(w.ListenPort)
	return &pipeline.Result{}
}
