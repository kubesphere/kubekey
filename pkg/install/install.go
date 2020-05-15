package install

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/cluster/etcd"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/container-engine/docker"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/plugins/network"
	"github.com/kubesphere/kubekey/pkg/plugins/storage"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func CreateCluster(clusterCfgFile string, logger *log.Logger, addons, pkg string, verbose bool) error {
	cfg, err := config.ParseClusterCfg(clusterCfgFile, addons, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}

	if err := preinstall.Prepare(&cfg.Spec, logger); err != nil {
		return errors.Wrap(err, "Failed to load kube binaries")
	}
	return Execute(executor.NewExecutor(&cfg.Spec, logger, verbose))
}

func ExecTasks(mgr *manager.Manager) error {
	createTasks := []manager.Task{
		{Task: preinstall.InitOS, ErrMsg: "Failed to download kube binaries"},
		{Task: docker.InstallerDocker, ErrMsg: "Failed to install docker"},
		{Task: images.PreDownloadImages, ErrMsg: "Failed to pre-download images"},
		{Task: etcd.GenerateEtcdCerts, ErrMsg: "Failed to generate etcd certs"},
		{Task: etcd.SyncEtcdCertsToMaster, ErrMsg: "Failed to sync etcd certs"},
		{Task: etcd.GenerateEtcdService, ErrMsg: "Failed to start etcd cluster"},
		{Task: kubernetes.GetClusterStatus, ErrMsg: "Failed to get cluster status"},
		{Task: kubernetes.SyncKubeBinaries, ErrMsg: "Failed to sync kube binaries"},
		{Task: kubernetes.InitKubernetesCluster, ErrMsg: "Failed to init kubernetes cluster"},
		{Task: network.DeployNetworkPlugin, ErrMsg: "Failed to deploy network plugin"},
		//{Task: kubernetes.GetJoinNodesCmd, ErrMsg: "Failed to get join cmd"},
		{Task: kubernetes.JoinNodesToCluster, ErrMsg: "Failed to join node"},
		{Task: storage.DeployStoragePlugins, ErrMsg: "Failed to deploy storage plugin"},
		{Task: kubesphere.DeployKubeSphere, ErrMsg: "Failed to deploy kubesphere"},
	}

	for _, step := range createTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	fmt.Printf("\n\033[1;36;40m%s\033[0m\n", "Successful.")
	return nil
}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}
