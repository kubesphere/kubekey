package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type Copy struct {
	BaseAction
	Src    string
	Dst    string
	Result ending.Result
}

func (c *Copy) Execute(vars vars.Vars) *ending.Result {
	fmt.Println(c.Dst, c.Src)
	return &ending.Result{}
}
