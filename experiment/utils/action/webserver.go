package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type WebServer struct {
	ListenPort    int
	ListenAddress string // 127.0.0.1, 0.0.0.0, *
	Status        string // start, restart, stop
}

func (w *WebServer) Execute(vars vars.Vars) *ending.Result {
	fmt.Println(w.ListenPort)
	return &ending.Result{}
}
