package scale

import (
	"encoding/json"
	"fmt"
	"github.com/pixiake/kubekey/pkg/cluster/kubernetes"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall"
	"github.com/pixiake/kubekey/pkg/config"
	"github.com/pixiake/kubekey/pkg/container-engine/docker"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ScaleCluster(clusterCfgFile string, logger *log.Logger, pkg string) error {
	cfg, err := config.ParseClusterCfg(clusterCfgFile, logger)
	if err != nil {
		return errors.Wrap(err, "failed to download cluster config")
	}

	out, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(out))
	if err := preinstall.Prepare(&cfg.Spec, logger); err != nil {
		return errors.Wrap(err, "failed to load kube binarys")
	}
	return NewExecutor(&cfg.Spec, logger).Execute()
}

func ExecTasks(mgr *manager.Manager) error {
	scaleTasks := []manager.Task{
		{Fn: preinstall.InitOS, ErrMsg: "failed to download kube binaries"},
		{Fn: docker.InstallerDocker, ErrMsg: "failed to install docker"},
		{Fn: kubernetes.SyncKubeBinaries, ErrMsg: "failed to sync kube binaries"},
		{Fn: kubernetes.ConfigureKubeletService, ErrMsg: "failed to sync kube binaries"},
		{Fn: kubernetes.GetJoinNodesCmd, ErrMsg: "failed to get join cmd"},
		{Fn: kubernetes.JoinNodesToCluster, ErrMsg: "failed to join node"},
	}

	for _, task := range scaleTasks {
		if err := task.Run(mgr); err != nil {
			return errors.Wrap(err, task.ErrMsg)
		}
	}
	return nil
}
