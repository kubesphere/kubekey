package tasks

import (
	"github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/experiment/utils/config"
	"github.com/kubesphere/kubekey/experiment/utils/pipeline"
	"github.com/kubesphere/kubekey/pkg/util/manager"
)

var (
	initClusterCmd   = "kubeadm init -f kubeadm.conf"
	getKubeConfigCmd = "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
	addNodeCmdTmpl   = "kubeadm join {{ .ApiServer }} --token {{ .Token }} --discovery-token-ca-cert-hash {{ .Hash }}"

	mgr         = manager.Manager{}
	InitCluster = pipeline.Task{
		Name:  "Init Cluster",
		Hosts: []v1alpha1.HostCfg{mgr.MasterNodes[0]},
		Action: &pipeline.Command{
			Cmd: initClusterCmd,
		},
		Env: nil,
		Vars: pipeline.Vars{
			"kubernetes": config.GetConfig().Cluster.Kubernetes.ClusterName,
		},
		Parallel:    false,
		Prepare:     &pipeline.Condition{Cond: true},
		IgnoreError: false,
	}

	GetKubeConfig = pipeline.Task{
		Name:    "Get KubeConfig",
		Hosts:   mgr.MasterNodes,
		Action:  &pipeline.Command{Cmd: getKubeConfigCmd},
		Prepare: &pipeline.Condition{Cond: true},
	}
)
