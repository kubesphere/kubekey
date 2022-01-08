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

package os

import (
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
)

type NodeConfigureNtpCheck struct {
	common.KubePrepare
}

func (n *NodeConfigureNtpCheck) PreCheck(_ connector.Runtime) (bool, error) {
	// skip when both NtpServers and Timezone was not set in cluster config
	if len(n.KubeConf.Cluster.System.NtpServers) == 0 && len(n.KubeConf.Cluster.System.Timezone) == 0 {
		return false, nil
	}

	return true, nil
}
