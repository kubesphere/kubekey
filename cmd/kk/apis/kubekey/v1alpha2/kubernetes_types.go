/*
 Copyright 2021 The KubeSphere Authors.

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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

// Kubernetes contains the configuration for the cluster
type Kubernetes struct {
	Type                   string   `yaml:"type" json:"type,omitempty"`
	Version                string   `yaml:"version" json:"version,omitempty"`
	ClusterName            string   `yaml:"clusterName" json:"clusterName,omitempty"`
	DNSDomain              string   `yaml:"dnsDomain" json:"dnsDomain,omitempty"`
	DisableKubeProxy       bool     `yaml:"disableKubeProxy" json:"disableKubeProxy,omitempty"`
	MasqueradeAll          bool     `yaml:"masqueradeAll" json:"masqueradeAll,omitempty"`
	MaxPods                int      `yaml:"maxPods" json:"maxPods,omitempty"`
	PodPidsLimit           int      `yaml:"podPidsLimit" json:"podPidsLimit,omitempty"`
	NodeCidrMaskSize       int      `yaml:"nodeCidrMaskSize" json:"nodeCidrMaskSize,omitempty"`
	NodeCidrMaskSizeIPv6   int      `yaml:"nodeCidrMaskSizeIPv6" json:"nodeCidrMaskSizeIPv6,omitempty"`
	ApiserverCertExtraSans []string `yaml:"apiserverCertExtraSans" json:"apiserverCertExtraSans,omitempty"`
	ProxyMode              string   `yaml:"proxyMode" json:"proxyMode,omitempty"`
	AutoRenewCerts         *bool    `yaml:"autoRenewCerts" json:"autoRenewCerts,omitempty"`
	// +optional
	Nodelocaldns             *bool                `yaml:"nodelocaldns" json:"nodelocaldns,omitempty"`
	ContainerManager         string               `yaml:"containerManager" json:"containerManager,omitempty"`
	ContainerRuntimeEndpoint string               `yaml:"containerRuntimeEndpoint" json:"containerRuntimeEndpoint,omitempty"`
	NodeFeatureDiscovery     NodeFeatureDiscovery `yaml:"nodeFeatureDiscovery" json:"nodeFeatureDiscovery,omitempty"`
	Kata                     Kata                 `yaml:"kata" json:"kata,omitempty"`
	ApiServerArgs            []string             `yaml:"apiserverArgs" json:"apiserverArgs,omitempty"`
	ControllerManagerArgs    []string             `yaml:"controllerManagerArgs" json:"controllerManagerArgs,omitempty"`
	SchedulerArgs            []string             `yaml:"schedulerArgs" json:"schedulerArgs,omitempty"`
	KubeletArgs              []string             `yaml:"kubeletArgs" json:"kubeletArgs,omitempty"`
	KubeProxyArgs            []string             `yaml:"kubeProxyArgs" json:"kubeProxyArgs,omitempty"`
	FeatureGates             map[string]bool      `yaml:"featureGates" json:"featureGates,omitempty"`
	KubeletConfiguration     runtime.RawExtension `yaml:"kubeletConfiguration" json:"kubeletConfiguration,omitempty"`
	KubeProxyConfiguration   runtime.RawExtension `yaml:"kubeProxyConfiguration" json:"kubeProxyConfiguration,omitempty"`
	Audit                    Audit                `yaml:"audit" json:"audit,omitempty"`
	NvidiaRuntime            *bool                `yaml:"nvidiaRuntime" json:"nvidiaRuntime,omitempty"`
}

// Kata contains the configuration for the kata in cluster
type Kata struct {
	Enabled *bool `yaml:"enabled" json:"enabled,omitempty"`
}

// NodeFeatureDiscovery contains the configuration for the node-feature-discovery in cluster
type NodeFeatureDiscovery struct {
	Enabled *bool `yaml:"enabled" json:"enabled,omitempty"`
}

// Audit contains the configuration for the kube-apiserver audit in cluster
type Audit struct {
	Enabled *bool `yaml:"enabled" json:"enabled,omitempty"`
}

// EnableNodelocaldns is used to determine whether to deploy nodelocaldns.
func (k *Kubernetes) EnableNodelocaldns() bool {
	if k.Nodelocaldns == nil {
		return true
	}
	return *k.Nodelocaldns
}

// EnableKataDeploy is used to determine whether to deploy kata.
func (k *Kubernetes) EnableKataDeploy() bool {
	if k.Kata.Enabled == nil {
		return false
	}
	return *k.Kata.Enabled
}

// EnableNodeFeatureDiscovery is used to determine whether to deploy node-feature-discovery.
func (k *Kubernetes) EnableNodeFeatureDiscovery() bool {
	if k.NodeFeatureDiscovery.Enabled == nil {
		return false
	}
	return *k.NodeFeatureDiscovery.Enabled
}

// EnableAutoRenewCerts is used to determine whether to enable AutoRenewCerts.
func (k *Kubernetes) EnableAutoRenewCerts() bool {
	if k.AutoRenewCerts == nil {
		return false
	}
	return *k.AutoRenewCerts
}

// EnableAudit is used to determine whether to enable kube-apiserver audit.
func (k *Kubernetes) EnableAudit() bool {
	if k.Audit.Enabled == nil {
		return false
	}
	return *k.AutoRenewCerts
}

// IsAtLeastV124 is used to determine whether the k8s version is greater than v1.24.
func (k *Kubernetes) IsAtLeastV124() bool {
	parsedVersion, err := versionutil.ParseGeneric(k.Version)
	if err != nil {
		return false
	}

	if parsedVersion.AtLeast(versionutil.MustParseSemantic("v1.24.0")) {
		return true
	}

	return false
}

// EnableNvidiaRuntime is used to determine whether to enable NVIDIA container runtime.
func (k *Kubernetes) EnableNvidiaRuntime() bool {
	if k.NvidiaRuntime == nil {
		return false
	}
	return *k.NvidiaRuntime
}
