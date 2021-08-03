package action

import (
	"fmt"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
)

type Template struct {
	BaseAction
	Dst  string
	Data map[string]interface{}
}

func (t *Template) Execute(vars pipeline.Vars) *pipeline.Result {
	fmt.Println(t.Data)
	return &pipeline.Result{}
}
