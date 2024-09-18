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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PipelineTemplateSpec defines the desired state of PipelineTemplate.
type PipelineTemplateSpec struct {
	// Project is storage for executable packages
	// +optional
	Project PipelineProject `json:"project,omitempty"`
	// InventoryRef is the node configuration for playbook
	// +optional
	InventoryRef *corev1.ObjectReference `json:"inventoryRef,omitempty"`
	// ConfigRef is the global variable configuration for playbook
	// +optional
	ConfigRef *corev1.ObjectReference `json:"configRef,omitempty"`
	// Tags is the tags of playbook which to execute
	// +optional
	Tags []string `json:"tags,omitempty"`
	// SkipTags is the tags of playbook which skip execute
	// +optional
	SkipTags []string `json:"skipTags,omitempty"`
	// If Debug mode is true, It will retain runtime data after a successful execution of Pipeline,
	// which includes task execution status and parameters.
	// +optional
	Debug bool `json:"debug,omitempty"`
	// when execute in kubernetes, pipeline will create ob or cornJob to execute.
	// +optional
	JobSpec PipelineJobSpec `json:"jobSpec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=pipelinetemplates,scope=Namespaced
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Time duration since creation of PipelineTemplate"
// +k8s:defaulter-gen=true

// PipelineTemplate is the Schema for the pipelinetemplates API
type PipelineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PipelineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// PipelineTemplateList contains a list of PipelineTemplate
type PipelineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineTemplate{}, &PipelineTemplateList{})
}
