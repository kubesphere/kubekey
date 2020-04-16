package install

import (
	"github.com/pixiake/kubekey/cluster/container-engine/docker"
	"github.com/pixiake/kubekey/cluster/kubernetes"
	"github.com/pixiake/kubekey/cluster/preinstall"
	"github.com/pixiake/kubekey/util/manager"
	"github.com/pixiake/kubekey/util/task"
	"github.com/pkg/errors"
)

func ExecTasks(mgr *manager.Manager) error {
	createTasks := []task.Task{
		{Fn: preinstall.InitOS, ErrMsg: "failed to download kube binaries"},
		{Fn: docker.InstallerDocker, ErrMsg: "failed to install docker"},
		{Fn: kubernetes.SyncKubeBinaries, ErrMsg: "failed to sync kube binaries"},
		{Fn: kubernetes.ConfigureKubeletService, ErrMsg: "failed to sync kube binaries"},
	}

	for _, step := range createTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}
	return nil
}
