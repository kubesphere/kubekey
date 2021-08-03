package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Copy struct {
	BaseAction
	Src    string
	Dst    string
	Result pipeline.Result
}

func (c *Copy) Execute(vars pipeline.Vars) *pipeline.Result {
	fmt.Println(c.Dst, c.Src)
	return &pipeline.Result{}
}
