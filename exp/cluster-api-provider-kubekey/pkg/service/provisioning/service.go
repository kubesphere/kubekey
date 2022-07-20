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

package provisioning

import (
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/provisioning/cloudinit"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/provisioning/commands"
)

type Service interface {
	RawBootstrapDataToProvisioningCommands(config []byte) ([]commands.Cmd, error)
}

func NewService(sshClient ssh.Interface, format bootstrapv1.Format) Service {
	switch format {
	case bootstrapv1.CloudConfig:
		return cloudinit.NewService(sshClient)
	default:
		return cloudinit.NewService(sshClient)
	}
}
