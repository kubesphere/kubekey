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
	"fmt"
	"strings"

	"github.com/jinzhu/copier"
	kubeyaml "sigs.k8s.io/yaml"

	infrabootstrapv1 "github.com/kubesphere/kubekey/v3/bootstrap/k3s/api/v1beta1"
)

// MarshalInitServerConfiguration marshals the ServerConfiguration object into a string.
func MarshalInitServerConfiguration(spec *infrabootstrapv1.K3sConfigSpec, token string) (string, error) {
	obj := spec.ServerConfiguration
	serverConfig := &K3sServerConfiguration{}
	if err := copier.Copy(serverConfig, obj.Database); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.Listener); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.Networking); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.KubernetesComponents); err != nil {
		return "", err
	}

	serverConfig.Token = token
	serverConfig.ClusterInit = *obj.Database.ClusterInit

	serverConfig.DisableCloudController = true
	serverConfig.KubeAPIServerArgs = append(obj.KubernetesProcesses.KubeAPIServerArgs, "anonymous-auth=true", getTLSCipherSuiteArg())
	serverConfig.KubeControllerManagerArgs = append(obj.KubernetesProcesses.KubeControllerManagerArgs, "cloud-provider=external")
	serverConfig.KubeSchedulerArgs = obj.KubernetesProcesses.KubeSchedulerArgs

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
		KubeletArgs:              obj.Agent.KubernetesAgentProcesses.KubeletArgs,
		KubeProxyArgs:            obj.Agent.KubernetesAgentProcesses.KubeProxyArgs,
	}

	b, err := kubeyaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJoinServerConfiguration marshals the join ServerConfiguration object into a string.
func MarshalJoinServerConfiguration(spec *infrabootstrapv1.K3sConfigSpec) (string, error) {
	obj := spec.ServerConfiguration
	serverConfig := &K3sServerConfiguration{}
	if err := copier.Copy(serverConfig, obj.Database); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.Listener); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.Networking); err != nil {
		return "", err
	}
	if err := copier.Copy(serverConfig, obj.KubernetesComponents); err != nil {
		return "", err
	}

	serverConfig.TokenFile = spec.Cluster.TokenFile
	serverConfig.Token = spec.Cluster.Token
	serverConfig.Server = spec.Cluster.Server

	serverConfig.DisableCloudController = true
	serverConfig.KubeAPIServerArgs = append(obj.KubernetesProcesses.KubeAPIServerArgs, "anonymous-auth=true", getTLSCipherSuiteArg())
	serverConfig.KubeControllerManagerArgs = append(obj.KubernetesProcesses.KubeControllerManagerArgs, "cloud-provider=external")
	serverConfig.KubeSchedulerArgs = obj.KubernetesProcesses.KubeSchedulerArgs

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
		KubeletArgs:              obj.Agent.KubernetesAgentProcesses.KubeletArgs,
		KubeProxyArgs:            obj.Agent.KubernetesAgentProcesses.KubeProxyArgs,
	}

	b, err := kubeyaml.Marshal(serverConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// MarshalJoinAgentConfiguration marshals the join AgentConfiguration object into a string.
func MarshalJoinAgentConfiguration(spec *infrabootstrapv1.K3sConfigSpec) (string, error) {
	obj := spec.AgentConfiguration
	agentConfig := &K3sAgentConfiguration{}
	if err := copier.Copy(agentConfig, obj.Node); err != nil {
		return "", err
	}
	if err := copier.Copy(agentConfig, obj.Networking); err != nil {
		return "", err
	}
	if err := copier.Copy(agentConfig, obj.Runtime); err != nil {
		return "", err
	}
	if err := copier.Copy(agentConfig, obj.KubernetesAgentProcesses); err != nil {
		return "", err
	}

	agentConfig.TokenFile = spec.Cluster.TokenFile
	agentConfig.Token = spec.Cluster.Token
	agentConfig.Server = spec.Cluster.Server

	b, err := kubeyaml.Marshal(agentConfig)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func getTLSCipherSuiteArg() string {
	ciphers := []string{
		// Modern Compatibility recommended configuration in
		// https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		"TLS_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384",
	}

	ciphersList := ""
	for _, cc := range ciphers {
		ciphersList += cc + ","
	}
	ciphersList = strings.TrimRight(ciphersList, ",")

	return fmt.Sprintf("tls-cipher-suites=%s", ciphersList)
}
