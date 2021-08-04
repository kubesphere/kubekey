package tasks

import (
	"github.com/kubesphere/kubekey/experiment/utils/action"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/ending"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
	"github.com/kubesphere/kubekey/experiment/utils/prepare"
	"github.com/kubesphere/kubekey/experiment/utils/vars"
	"github.com/pkg/errors"
	"strings"
)

type getClusterAction struct {
	action.BaseAction
}

func (g *getClusterAction) Execute(vars vars.Vars) *ending.Result {
	var clusterIsExist bool
	output, _, _, _ := g.Manager.Runner.Cmd("sudo -E /bin/sh -c \"[ -f /etc/kubernetes/admin.conf ] && echo 'Cluster already exists.' || echo 'Cluster will be created.'\"", true)
	if strings.Contains(output, "Cluster will be created") {
		clusterIsExist = false
	} else {
		clusterIsExist = true
	}
	g.Pool.Set("IsExist", clusterIsExist)
	return ending.NewResult()
}

type generateCfgAction struct {
	action.BaseAction
}

func (g *generateCfgAction) Execute(vars vars.Vars) *ending.Result {
	exist, ok := g.Pool.GetMustBool("IsExist")
	if !ok {
		return ending.NewResultWithErr(errors.New("failed to get var that in the Pool"))
	}
	if exist {

	}
	return ending.NewResult()
}

func NewGetClusterStatusModule() *pipeline.TaskModule {
	mgr = config.GetManager()

	getClusterTask := pipeline.Task{
		Name:    "getClusterTask",
		Hosts:   mgr.MasterNodes,
		Prepare: new(prepare.OnlyFirstMaster),
		Action:  new(getClusterAction),
	}

	generateTask := pipeline.Task{
		Name:   "GenerateConfigTask",
		Hosts:  mgr.K8sNodes,
		Action: new(generateCfgAction),
	}

	tasks := []pipeline.Task{
		getClusterTask,
		generateTask,
	}

	return pipeline.NewTaskModule(tasks)
}
