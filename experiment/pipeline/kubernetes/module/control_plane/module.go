package control_plane

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/experiment/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/core/action"
	"github.com/kubesphere/kubekey/experiment/core/pipeline"
	"github.com/kubesphere/kubekey/experiment/core/prepare"
	"github.com/kubesphere/kubekey/experiment/core/vars"
	"github.com/pkg/errors"
	"strings"
)

type getClusterAction struct {
	action.BaseAction
}

func (g *getClusterAction) Execute(vars vars.Vars) error {
	var clusterIsExist bool
	output, _, _, _ := g.Manager.Runner.Cmd("sudo -E /bin/sh -c \"[ -f /etc/kubernetes/admin.conf ] && echo 'Cluster already exists.' || echo 'Cluster will be created.'\"", true)
	if strings.Contains(output, "Cluster will be created") {
		clusterIsExist = false
	} else {
		clusterIsExist = true
	}
	g.Cache.Set("IsExist", clusterIsExist)
	return nil
}

type generateCfgAction struct {
	action.BaseAction
}

func (g *generateCfgAction) Execute(vars vars.Vars) error {
	exist, ok := g.Cache.GetMustBool("IsExist")
	if !ok {
		return errors.New("failed to get var that in the Pool")
	}
	if exist {

	}
	return nil
}

type GetClusterStatusModule struct {
	pipeline.DefaultTaskModule
}

func (g *GetClusterStatusModule) Init() {
	g.Name = "GetClusterStatus"

	getClusterTask := pipeline.Task{
		Name:    "getClusterTask",
		Hosts:   []kubekeyapiv1alpha1.HostCfg{g.Runtime.MasterNodes[0]},
		Prepare: new(prepare.OnlyFirstMaster),
		Action:  new(getClusterAction),
	}

	generateTask := pipeline.Task{
		Name:   "GenerateConfigTask",
		Hosts:  g.Runtime.MasterNodes,
		Action: new(generateCfgAction),
	}

	g.Tasks = []pipeline.Task{
		getClusterTask,
		generateTask,
	}
}
