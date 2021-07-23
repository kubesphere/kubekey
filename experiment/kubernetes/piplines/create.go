package piplines

import (
	"github.com/kubesphere/kubekey/experiment/kubernetes/control-plane/tasks"
	"github.com/kubesphere/kubekey/experiment/utils/pipline"
)

var (
	CreateClusterPipline = pipline.Pipline{TaskList: []pipline.Task{
       tasks.InitCluster,
       tasks.GetKubeConfig,
	}}
)