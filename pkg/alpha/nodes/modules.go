/*
 Copyright 2022 The KubeSphere Authors.

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

package nodes

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/prepare"
	"github.com/kubesphere/kubekey/pkg/core/task"
	"github.com/kubesphere/kubekey/pkg/kubernetes"
)

type UpgradeNodesModule struct {
	common.KubeModule
}

func (p *UpgradeNodesModule) Init() {
	p.Name = "UpgradeNodesModule"
	p.Desc = "Upgrade cluster on all nodes"

	upgradeKubeMaster := &task.RemoteTask{
		Name:     "UpgradeClusterOnMaster",
		Desc:     "Upgrade cluster on master",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(kubernetes.NotEqualPlanVersion),
		Action:   &kubernetes.UpgradeKubeMaster{ModuleName: p.Name},
		Parallel: false,
	}

	cluster := kubernetes.NewKubernetesStatus()
	p.PipelineCache.GetOrSet(common.ClusterStatus, cluster)

	clusterStatus := &task.RemoteTask{
		Name:     "GetClusterStatus",
		Desc:     "Get kubernetes cluster status",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(kubernetes.NotEqualPlanVersion),
		Action:   new(kubernetes.GetClusterStatus),
		Parallel: false,
	}

	upgradeNodes := &task.RemoteTask{
		Name:  "UpgradeClusterOnWorker",
		Desc:  "Upgrade cluster on worker",
		Hosts: p.Runtime.GetHostsByRole(common.Worker),
		Prepare: &prepare.PrepareCollection{
			new(kubernetes.NotEqualPlanVersion),
			new(common.OnlyWorker),
		},
		Action:   &kubernetes.UpgradeKubeWorker{ModuleName: p.Name},
		Parallel: false,
	}

	reconfigureDNS := &task.RemoteTask{
		Name:     "ReconfigureCoreDNS",
		Desc:     "Reconfigure CoreDNS",
		Hosts:    p.Runtime.GetHostsByRole(common.Master),
		Action:   &kubernetes.ReconfigureDNS{ModuleName: p.Name},
		Parallel: false,
	}

	p.Tasks = []task.Interface{
		upgradeKubeMaster,
		clusterStatus,
		upgradeNodes,
		reconfigureDNS,
	}
}
