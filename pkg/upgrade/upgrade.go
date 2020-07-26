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

package upgrade

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/kubesphere"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func UpgradeCluster(clusterCfgFile, k8sVersion, ksVersion string, logger *log.Logger, ksEnabled, verbose bool) error {
	currentDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "Faild to get current dir")
	}
	if err := util.CreateDir(fmt.Sprintf("%s/kubekey", currentDir)); err != nil {
		return errors.Wrap(err, "Failed to create work dir")
	}

	cfg, err := config.ParseClusterCfg(clusterCfgFile, k8sVersion, ksVersion, ksEnabled, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}

	return Execute(executor.NewExecutor(&cfg.Spec, logger, verbose, true))
}

func ExecTasks(mgr *manager.Manager) error {
	scaleTasks := []manager.Task{
		{Task: GetClusterInfo, ErrMsg: "Failed to get cluster info"},
		//{Task: preinstall.InitOS, ErrMsg: "Failed to download kube binaries"},
		//{Task: preinstall.PrePullImages, ErrMsg: "Failed to pre-pull images"},
		{Task: kubernetes.GetCurrentVersions, ErrMsg: "Failed to get current version"},
		{Task: kubernetes.UpgradeKubeCluster, ErrMsg: "Failed to upgrade kube cluster"},
		{Task: kubesphere.DeployKubeSphere, ErrMsg: "Failed to upgrade kubesphere"},
	}

	for _, step := range scaleTasks {
		if err := step.Run(mgr); err != nil {
			return errors.Wrap(err, step.ErrMsg)
		}
	}

	if mgr.KsEnable {
		mgr.Logger.Infoln(`Upgrading is complete.

Please check the result using the command:

       kubectl logs -n kubesphere-system $(kubectl get pod -n kubesphere-system -l app=ks-install -o jsonpath='{.items[0].metadata.name}') -f

`)
	} else {
		mgr.Logger.Infoln("Congradulations! Upgrade cluster is successful.")
	}

	return nil
}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}
