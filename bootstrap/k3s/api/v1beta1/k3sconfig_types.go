/*
Copyright 2022.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
)

// K3sConfigSpec defines the desired state of K3sConfig
type K3sConfigSpec struct {
	// Files specifies extra files to be passed to user_data upon creation.
	// +optional
	Files []bootstrapv1.File `json:"files,omitempty"`

	// Cluster defines the k3s cluster Options.
	Cluster *Cluster `json:"cluster,omitempty"`

	// ServerConfiguration defines the k3s server configuration.
	// +optional
	ServerConfiguration *ServerConfiguration `json:"serverConfiguration,omitempty"`

	// AgentConfiguration defines the k3s agent configuration.
	// +optional
	AgentConfiguration *AgentConfiguration `json:"agentConfiguration,omitempty"`

	// PreK3sCommands specifies extra commands to run before k3s setup runs
	// +optional
	PreK3sCommands []string `json:"preK3sCommands,omitempty"`

	// PostK3sCommands specifies extra commands to run after k3s setup runs
	// +optional
	PostK3sCommands []string `json:"postK3sCommands,omitempty"`

	// Version specifies the k3s version
	// +optional
	Version string `json:"version,omitempty"`
}

// K3sConfigStatus defines the observed state of K3sConfig
type K3sConfigStatus struct {
	// Ready indicates the BootstrapData field is ready to be consumed
	Ready bool `json:"ready,omitempty"`

	BootstrapData []byte `json:"bootstrapData,omitempty"`

	// DataSecretName is the name of the secret that stores the bootstrap data script.
	// +optional
	DataSecretName *string `json:"dataSecretName,omitempty"`

	// FailureReason will be set on non-retryable errors
	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// FailureMessage will be set on non-retryable errors
	// +optional
	FailureMessage string `json:"failureMessage,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the K3sConfig.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=k3sconfigs,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels['cluster\\.x-k8s\\.io/cluster-name']",description="Cluster"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of K3sConfig"

// K3sConfig is the Schema for the k3sConfigs API
type K3sConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K3sConfigSpec   `json:"spec,omitempty"`
	Status K3sConfigStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (c *K3sConfig) GetConditions() clusterv1.Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *K3sConfig) SetConditions(conditions clusterv1.Conditions) {
	c.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// K3sConfigList contains a list of K3sConfig
type K3sConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3sConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3sConfig{}, &K3sConfigList{})
}
