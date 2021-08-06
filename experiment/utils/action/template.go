package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
)

type Template struct {
	BaseAction
	Dst  string
	Data map[string]interface{}
}

func (t *Template) Execute(vars vars.Vars) error {
	fmt.Println(t.Data)
	return nil
}
