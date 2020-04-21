package install

import (
	"github.com/pixiake/kubekey/pkg/cluster/container-engine/docker"
	"github.com/pixiake/kubekey/pkg/cluster/etcd"
	"github.com/pixiake/kubekey/pkg/cluster/kubernetes"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall"
	"github.com/pixiake/kubekey/pkg/plugins/network"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/task"
	"github.com/pkg/errors"
)

func ExecTasks(mgr *manager.Manager) error {
	createTasks := []task.Task{
		{Fn: preinstall.InitOS, ErrMsg: "failed to download kube binaries"},
		{Fn: docker.InstallerDocker, ErrMsg: "failed to install docker"},
		{Fn: kubernetes.SyncKubeBinaries, ErrMsg: "failed to sync kube binaries"},
		{Fn: etcd.GenerateEtcdCerts, ErrMsg: "failed to generate etcd certs"},
		{Fn: etcd.SyncEtcdCertsToMaster, ErrMsg: "failed to sync etcd certs"},
		{Fn: etcd.GenerateEtcdService, ErrMsg: "failed to start etcd cluster"},
		{Fn: kubernetes.ConfigureKubeletService, ErrMsg: "failed to sync kube binaries"},
		{Fn: kubernetes.InitKubernetesCluster, ErrMsg: "failed to init kubernetes cluster"},
		{Fn: network.DeployNetworkPlugin, ErrMsg: "failed to deploy network plugin"},
		{Fn: kubernetes.GetJoinNodesCmd, ErrMsg: "failed to get join cmd"},
		{Fn: kubernetes.JoinNodesToCluster, ErrMsg: "failed to join node"},
	}

	for _, step := range createTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}
	return nil
}
