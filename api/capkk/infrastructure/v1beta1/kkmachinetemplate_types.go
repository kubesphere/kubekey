/*
Copyright 2024 The KubeSphere Authors.

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
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// KKMachineTemplateSpec defines the desired state of KKMachineTemplate.
type KKMachineTemplateSpec struct {
	Template KKMachineTemplateResource `json:"template"`
}

// KKMachineTemplateResource describes the data needed to create a KKMachine from a template.
type KKMachineTemplateResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1beta1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired behavior of the machine.
	Spec KKMachineSpec `json:"spec"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kkmachinetemplates,scope=Namespaced,categories=cluster-api,shortName=kkmt
// +kubebuilder:storageversion
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1beta1"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of KKMachineTemplate"
// +k8s:defaulter-gen=true

// KKMachineTemplate is the Schema for the kkmachinetemplates API
type KKMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KKMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// KKMachineTemplateList contains a list of KKMachineTemplate
type KKMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKMachineTemplate{}, &KKMachineTemplateList{})
}
