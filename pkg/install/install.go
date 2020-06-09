/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func CreateCluster(clusterCfgFile, k8sVersion, ksVersion string, logger *log.Logger, all, verbose bool) error {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Faild to get current dir")
	}
	if err := util.CreateDir(fmt.Sprintf("%s/kubekey", currentDir)); err != nil {
		return errors.Wrap(err, "Failed to create work dir")
	}

	cfg, err := config.ParseClusterCfg(clusterCfgFile, k8sVersion, ksVersion, all, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}

	return Execute(executor.NewExecutor(&cfg.Spec, logger, verbose))
}

func ExecTasks(mgr *manager.Manager) error {
	createTasks := []manager.Task{
		{Task: preinstall.Precheck, ErrMsg: "Failed to precheck"},
		{Task: preinstall.InitOS, ErrMsg: "Failed to download kube binaries"},
		{Task: docker.InstallerDocker, ErrMsg: "Failed to install docker"},
		{Task: images.PreDownloadImages, ErrMsg: "Failed to pre-download images"},
		{Task: etcd.GenerateEtcdCerts, ErrMsg: "Failed to generate etcd certs"},
		{Task: etcd.SyncEtcdCertsToMaster, ErrMsg: "Failed to sync etcd certs"},
		{Task: etcd.GenerateEtcdService, ErrMsg: "Failed to create etcd service"},
		{Task: etcd.SetupEtcdCluster, ErrMsg: "Failed to start etcd cluster"},
		{Task: etcd.RefreshEtcdConfig, ErrMsg: "Failed to refresh etcd configuration"},
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

	mgr.Logger.Infoln("Congradulations! Installation is successful.")

	return nil
}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}
