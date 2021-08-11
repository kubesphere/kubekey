package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/core/vars"
)

type Copy struct {
	BaseAction
	Src string
	Dst string
}

func (c *Copy) Execute(vars vars.Vars) error {
	fmt.Println(c.Dst, c.Src)
	return nil
}
