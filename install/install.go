package install

import (
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/cluster/container-engine/docker"
	"github.com/pixiake/kubekey/util/state"
	"github.com/pixiake/kubekey/util/task"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateCluster(logger *log.Logger, clusterCfgFile string, addons string, pkg string) error {
	cfg := kubekeyapi.GetClusterCfg(clusterCfgFile)
	//installer.NewInstaller(cluster, logger)
	return NewInstaller(cfg, logger).Install()
}

func ExecTasks(s *state.State) error {
	createTasks := []task.Task{
		{Fn: docker.InstallerDocker, ErrMsg: "failed to download kube binaries"},
	}

	for _, step := range createTasks {
		if err := step.Run(s); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}
	return nil
}
