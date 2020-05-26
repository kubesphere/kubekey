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

package scale

import (
	"fmt"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes"
	"github.com/kubesphere/kubekey/pkg/cluster/preinstall"
	"github.com/kubesphere/kubekey/pkg/container-engine/docker"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

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
