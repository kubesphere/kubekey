package action

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
	"github.com/kubesphere/kubekey/pkg/util"
	"text/template"
)

type Command struct {
	mgr    *config.Manager
	Cmd    string
	Retry  int
	Result pipeline.Result
}

func (c *Command) Init(mgr *config.Manager) {
	c.mgr = mgr
}

func (c *Command) Execute(node *kubekeyapiv1alpha1.HostCfg, vars pipeline.Vars) *pipeline.Result {
	res := pipeline.NewResult()
	defer res.SetEndTime()

	nodeMap, err := config.GetNodeMap(node.Name)
	if err != nil {
		res.ErrResult(err)
		return res
	}

	// todo: best way to merge the nodeMap and the vars map
	nodeMap.Range(func(key, value interface{}) bool {
		if _, ok := vars[key.(string)]; !ok {
			vars[key.(string)] = value
		}
		return true
	})

	cmdTmpl := template.Must(template.New("").Parse(c.Cmd))
	cmd, err := util.Render(cmdTmpl, vars)
	if err != nil {
		res.ErrResult(err)
		return res
	}

	fmt.Println(cmd)
	// todo: run cmd. maybe need rewrite the runner
	return res
}
