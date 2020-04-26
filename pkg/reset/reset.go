package reset

import (
	"encoding/json"
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/cluster/preinstall"
	"github.com/pixiake/kubekey/pkg/config"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateCluster(clusterCfgFile string, logger *log.Logger, addons string, pkg string) error {
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
	resetTasks := []manager.Task{
		{Task: ResetKubeCluster, ErrMsg: "failed to reset kube cluster"},
		{Task: ResetEtcdCluster, ErrMsg: "failed to clean etcd files"},
	}

	for _, step := range resetTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}
	return nil
}

func ResetKubeCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Reset cluster")

	return mgr.RunTaskOnK8sNodes(resetKubeCluster, true)
}

func resetKubeCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	_, err := mgr.Runner.RunCmd("sudo -E /bin/sh -c \"/user/local/bin/kubeadm reset -f\"")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to reset kube cluster")
	}
	return nil
}

var etcdFiles = []string{"/usr/local/bin/etcd", "/etc/ssl/etcd/ssl", "/var/lib/etcd", "/etc/etcd.env", "/etc/systemd/system/etcd.service"}

func ResetEtcdCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Reset cluster")

	return mgr.RunTaskOnEtcdNodes(resetKubeCluster, false)
}

func resetEtcdCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	_, err := mgr.Runner.RunCmd("sudo -E /bin/sh -c \"systemctl stop etcd\"")
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to reset etcd cluster")
	}

	for _, file := range etcdFiles {
		_, err := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"rm -rf %s\"", file))
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("failed to clean etcd files: %s", file))
		}
	}
	return nil
}
