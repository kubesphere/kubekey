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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K3sConfigTemplateSpec defines the desired state of K3sConfigTemplate
type K3sConfigTemplateSpec struct {
	Template K3sConfigTemplateResource `json:"template"`
}

// K3sConfigTemplateResource defines the Template structure
type K3sConfigTemplateResource struct {
	Spec K3sConfigSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=k3sconfigtemplates,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of K3sConfigTemplate"

// K3sConfigTemplate is the Schema for the k3sconfigtemplates API
type K3sConfigTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec K3sConfigTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// K3sConfigTemplateList contains a list of K3sConfigTemplate
type K3sConfigTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3sConfigTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3sConfigTemplate{}, &K3sConfigTemplateList{})
}
