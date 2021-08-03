package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
	"github.com/kubesphere/kubekey/pkg/util"
	"text/template"
)

type Command struct {
	BaseAction
	Cmd    string
	Print  bool
	Result pipeline.Result
}

func (c *Command) Execute(vars pipeline.Vars) *pipeline.Result {
	res := pipeline.NewResult()
	defer res.SetEndTime()

	nodeMap, err := config.GetNodeMap(c.Manager.Runner.Host.Name)
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

	stdout, stderr, code, err := c.Manager.Runner.Cmd(cmd, c.Print)
	if err != nil {
		res.ErrResult(err)
		return res
	}

	res.NormalResult(code, stdout, stderr)
	return res
}
