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
		return errors.Wrap(err, "failed to download cluster config")
	}

	//out, _ := json.MarshalIndent(cfg, "", "  ")
	//fmt.Println(string(out))
	if err := preinstall.Prepare(&cfg.Spec, logger); err != nil {
		return errors.Wrap(err, "failed to load kube binarys")
	}
	return NewExecutor(&cfg.Spec, logger, verbose).Execute()
}

func ExecTasks(mgr *manager.Manager) error {
	scaleTasks := []manager.Task{
		{Task: preinstall.InitOS, ErrMsg: "failed to download kube binaries"},
		{Task: docker.InstallerDocker, ErrMsg: "failed to install docker"},
		{Task: kubernetes.SyncKubeBinaries, ErrMsg: "failed to sync kube binaries"},
		//{Task: kubernetes.ConfigureKubeletService, ErrMsg: "failed to sync kube binaries"},
		//{Task: kubernetes.GetJoinNodesCmd, ErrMsg: "failed to get join cmd"},
		{Task: kubernetes.JoinNodesToCluster, ErrMsg: "failed to join node"},
	}

	for _, task := range scaleTasks {
		if err := task.Run(mgr); err != nil {
			return errors.Wrap(err, task.ErrMsg)
		}
	}

	fmt.Printf("\n\033[1;36;40m%s\033[0m\n", "Successful.")
	return nil
}
