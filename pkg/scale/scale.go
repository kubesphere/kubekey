package scale

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/container-engine/docker"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ScaleCluster(clusterCfgFile string, logger *log.Logger, pkg string, verbose bool) error {
	cfg, err := config.ParseClusterCfg(clusterCfgFile, "", logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}

	//output, _ := json.MarshalIndent(cfg, "", "  ")
	//fmt.Println(string(output))
	if err := preinstall.Prepare(&cfg.Spec, logger); err != nil {
		return errors.Wrap(err, "Failed to load kube binarys")
	}
	return NewExecutor(&cfg.Spec, logger, verbose).Execute()
}

func ExecTasks(mgr *manager.Manager) error {
	scaleTasks := []manager.Task{
		{Task: preinstall.InitOS, ErrMsg: "Failed to download kube binaries"},
		{Task: docker.InstallerDocker, ErrMsg: "Failed to install docker"},
		{Task: kubernetes.SyncKubeBinaries, ErrMsg: "Failed to sync kube binaries"},
		//{Task: kubernetes.ConfigureKubeletService, ErrMsg: "Failed to sync kube binaries"},
		//{Task: kubernetes.GetJoinNodesCmd, ErrMsg: "Failed to get join cmd"},
		{Task: kubernetes.JoinNodesToCluster, ErrMsg: "Failed to join node"},
	}

	for _, task := range scaleTasks {
		if err := task.Run(mgr); err != nil {
			return errors.Wrap(err, task.ErrMsg)
		}
	}

	fmt.Printf("\n\033[1;36;40m%s\033[0m\n", "Successful.")
	return nil
}
