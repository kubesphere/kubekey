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

package k8e

import (
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
)

type NodeInCluster struct {
	common.KubePrepare
	Not bool
}

func (n *NodeInCluster) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := n.PipelineCache.Get(common.ClusterStatus); ok {
		cluster := v.(*K8eStatus)
		var versionOk bool
		if res, ok := cluster.NodesInfo[host.GetName()]; ok && res != "" {
			versionOk = true
		}
		_, ipOk := cluster.NodesInfo[host.GetInternalAddress()]
		if n.Not {
			return !(versionOk || ipOk), nil
		}
		return versionOk || ipOk, nil
	}
	return false, errors.New("get k8e cluster status by pipeline cache failed")
}

type ClusterIsExist struct {
	common.KubePrepare
	Not bool
}

func (c *ClusterIsExist) PreCheck(_ connector.Runtime) (bool, error) {
	if exist, ok := c.PipelineCache.GetMustBool(common.ClusterExist); ok {
		if c.Not {
			return !exist, nil
		}
		return exist, nil
	} else {
		return false, errors.New("get k8e cluster status by pipeline cache failed")
	}
}
