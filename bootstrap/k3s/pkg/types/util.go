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

package types

import (
	"github.com/jinzhu/copier"
	kubeyaml "sigs.k8s.io/yaml"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
)

// MarshalInitServerConfiguration marshals the ServerConfiguration object into a string.
func MarshalInitServerConfiguration(spec *infrabootstrapv1.K3sConfigSpec, token string) (string, error) {
	obj := spec.ServerConfiguration
	serverConfig := &K3sServerConfiguration{}
	if err := copier.Copy(serverConfig, obj); err != nil {
		return "", err
	}

	serverConfig.Token = token

	serverConfig.CloudInit = spec.ServerConfiguration.Database.ClusterInit

	serverConfig.K3sAgentConfiguration = K3sAgentConfiguration{
		NodeName:                 obj.Agent.Node.NodeName,
		NodeLabels:               obj.Agent.Node.NodeLabels,
		NodeTaints:               obj.Agent.Node.NodeTaints,
		SeLinux:                  obj.Agent.Node.SeLinux,
		LBServerPort:             obj.Agent.Node.LBServerPort,
		DataDir:                  obj.Agent.Node.DataDir,
		ContainerRuntimeEndpoint: obj.Agent.Runtime.ContainerRuntimeEndpoint,
		PauseImage:               obj.Agent.Runtime.PauseImage,
		PrivateRegistry:          obj.Agent.Runtime.PrivateRegistry,
		NodeIP:                   obj.Agent.Networking.NodeIP,
		NodeExternalIP:           obj.Agent.Networking.NodeExternalIP,
		ResolvConf:               obj.Agent.Networking.ResolvConf,
	}

	b, err := kubeyaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJoinServerConfiguration marshals the join ServerConfiguration object into a string.
func MarshalJoinServerConfiguration(obj *infrabootstrapv1.ServerConfiguration) (string, error) {
	serverConfig := &K3sServerConfiguration{}
	if err := copier.Copy(serverConfig, obj); err != nil {
		return "", err
	}

	serverConfig.K3sAgentConfiguration = K3sAgentConfiguration{
		NodeName:                 obj.Agent.Node.NodeName,
		NodeLabels:               obj.Agent.Node.NodeLabels,
		NodeTaints:               obj.Agent.Node.NodeTaints,
		SeLinux:                  obj.Agent.Node.SeLinux,
		LBServerPort:             obj.Agent.Node.LBServerPort,
		DataDir:                  obj.Agent.Node.DataDir,
		ContainerRuntimeEndpoint: obj.Agent.Runtime.ContainerRuntimeEndpoint,
		PauseImage:               obj.Agent.Runtime.PauseImage,
		PrivateRegistry:          obj.Agent.Runtime.PrivateRegistry,
		NodeIP:                   obj.Agent.Networking.NodeIP,
		NodeExternalIP:           obj.Agent.Networking.NodeExternalIP,
		ResolvConf:               obj.Agent.Networking.ResolvConf,
	}

	b, err := kubeyaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJoinAgentConfiguration marshals the join AgentConfiguration object into a string.
func MarshalJoinAgentConfiguration(obj *infrabootstrapv1.AgentConfiguration) (string, error) {
	serverConfig := &K3sAgentConfiguration{}
	if err := copier.Copy(serverConfig, obj); err != nil {
		return "", err
	}

	b, err := kubeyaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
