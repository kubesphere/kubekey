package piplines

import (
	"github.com/kubesphere/kubekey/experiment/kubernetes/control-plane/tasks"
	"github.com/kubesphere/kubekey/experiment/utils/pipline"
)

var (
	CreateClusterPipeline = pipline.Pipeline{TaskList: []pipline.Task{
		tasks.InitCluster,
		tasks.GetKubeConfig,
	}}
)
