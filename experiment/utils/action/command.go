package action

import (
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
	"github.com/kubesphere/kubekey/pkg/util"
	"text/template"
)

type Command struct {
	BaseAction
	Cmd   string
	Print bool
}

func (c *Command) Execute(vars vars.Vars) error {

	nodeMap, err := config.GetNodeMap(c.Manager.Runner.Host.Name)
	if err != nil {
		return err
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
		return err
	}

	_, _, _, err = c.Manager.Runner.Cmd(cmd, c.Print)
	if err != nil {
		return err
	}

	return nil
}
