package pipline

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/pkg/util"
	"text/template"
)

type Action interface {
	Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) (result *Result)
}

type Command struct {
	Cmd    string
	Result Result
}

type Copy struct {
	Src    string
	Dst    string
	Result Result
}

type Template struct {
	Dst  string
	Data map[string]interface{}
}

type WebServer struct {
	ListenPort    int
	ListenAddress string // 127.0.0.1, 0.0.0.0, *
	Status        string // start, restart, stop
}

type Func struct {
	function func(vars *Vars) *Result
}

func (f *Func) Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) *Result {
	return f.function(vars)
}

func (a *Command) Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) *Result {
	res := NewResult()
	defer res.SetEndTime()

	nodeMap, err := config.GetNodeMap(node.Name)
	if err != nil {
		res.ErrResult(err)
		return res
	}

	// todo: merge the nodeMap and the vars map
	cmdTmpl := template.Must(template.New("").Parse(a.Cmd))
	cmd, err := util.Render(cmdTmpl, nodeMap)
	if err != nil {
		res.ErrResult(err)
		return res
	}

	fmt.Println(cmd)
	return res
}

func (c *Copy) Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) *Result {
	fmt.Println(c.Dst, c.Src)
	return &Result{}
}

func (t *Template) Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) *Result {
	fmt.Println(t.Data)
	return &Result{}
}

func (w *WebServer) Execute(node kubekeyapiv1alpha1.HostCfg, vars *Vars) *Result {
	fmt.Println(w.ListenPort)
	return &Result{}
}
