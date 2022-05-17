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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AddonSpec defines the desired state of Addon
type AddonSpec struct {
	// ClusterName is the name of the Cluster
	ClusterName string `json:"clusterName"`

	// Components is the list of components to be installed
	Components []Component `json:"components,omitempty"`
}

type Component struct {
	// Provisioner is the name of the provisioner to use
	// Internal: "kubesphere.io/calico", "kubesphere.io/kubesphere"
	// external: "helm", "local"
	Provisioner string `json:"provisioner"`

	// Parameters is a map of parameters that's only for kubesphere.io provisioner
	Parameters map[string]string `json:"parameters,omitempty"`

	// Chart that's only for helm provisioner
	Chart Chart `json:"chart,omitempty"`

	// Yaml that's only for local provisioner
	Yaml Yaml `json:"yaml,omitempty"`
}

type Chart struct {
	Name       string   `yaml:"name" json:"name,omitempty"`
	Repo       string   `yaml:"repo" json:"repo,omitempty"`
	Path       string   `yaml:"path" json:"path,omitempty"`
	Version    string   `yaml:"version" json:"version,omitempty"`
	ValuesFile string   `yaml:"valuesFile" json:"valuesFile,omitempty"`
	Values     []string `yaml:"values" json:"values,omitempty"`
}

type Yaml struct {
	Path []string `yaml:"path" json:"path,omitempty"`
}

// AddonStatus defines the observed state of Addon
type AddonStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Addon is the Schema for the addons API
type Addon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AddonSpec   `json:"spec,omitempty"`
	Status AddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AddonList contains a list of Addon
type AddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Addon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Addon{}, &AddonList{})
}
