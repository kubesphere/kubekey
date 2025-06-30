/*
Copyright 2023 The KubeSphere Authors.

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
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// BuiltinsProjectAnnotation use builtins project of KubeKey
	BuiltinsProjectAnnotation = "kubekey.kubesphere.io/builtins-project"

	// PlaybookCompletedFinalizer will be removed after the Playbook is completed.
	PlaybookCompletedFinalizer = "kubekey.kubesphere.io/playbook-completed"
)

// PlaybookPhase of Playbook
type PlaybookPhase string

const (
	// PlaybookPhasePending of Playbook. Playbook has created but not deal
	PlaybookPhasePending PlaybookPhase = "Pending"
	// PlaybookPhaseRunning of Playbook. deal Playbook.
	PlaybookPhaseRunning PlaybookPhase = "Running"
	// PlaybookPhaseFailed of Playbook. once Task run failed.
	PlaybookPhaseFailed PlaybookPhase = "Failed"
	// PlaybookPhaseSucceed of Playbook. all Tasks run success.
	PlaybookPhaseSucceeded PlaybookPhase = "Succeeded"
)

// PlaybookFailedReason is the reason why a Playbook failed.
type PlaybookFailedReason string

const (
	// PlaybookFailedReasonUnknown is the default failed reason.
	PlaybookFailedReasonUnknown PlaybookFailedReason = "unknown"
	// PlaybookFailedReasonPodFailed pod exec failed.
	PlaybookFailedReasonPodFailed PlaybookFailedReason = "pod executor failed"
	// PlaybookFailedReasonTaskFailed task exec failed.
	PlaybookFailedReasonTaskFailed PlaybookFailedReason = "task executor failed"
)

// PlaybookSpec defines the desired state of Playbook.
type PlaybookSpec struct {
	// Project is storage for executable packages
	// +optional
	Project PlaybookProject `json:"project,omitempty"`
	// Playbook which to execute.
	Playbook string `json:"playbook"`
	// InventoryRef is the node configuration for playbook
	// +optional
	InventoryRef *corev1.ObjectReference `json:"inventoryRef,omitempty"`
	// Config is the global variable configuration for playbook
	// +optional
	Config Config `json:"config,omitempty"`
	// Tags is the tags of playbook which to execute
	// +optional
	Tags []string `json:"tags,omitempty"`
	// SkipTags is the tags of playbook which skip execute
	// +optional
	SkipTags []string `json:"skipTags,omitempty"`
	// Volumes in job pod.
	// +optional
	Volumes []corev1.Volume `json:"workVolume,omitempty"`
	// VolumeMounts in job pod.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
}

// PlaybookProject respect which playbook store.
type PlaybookProject struct {
	// Addr is the storage for executable packages (in Ansible file format).
	// When starting with http or https, it will be obtained from a Git repository.
	// When starting with file path, it will be obtained from the local path.
	// +optional
	Addr string `json:"addr,omitempty"`
	// Name is the project name base project
	// +optional
	Name string `json:"name,omitempty"`
	// Branch is the git branch of the git Addr.
	// +optional
	Branch string `json:"branch,omitempty"`
	// Tag is the git branch of the git Addr.
	// +optional
	Tag string `json:"tag,omitempty"`
	// InsecureSkipTLS skip tls or not when git addr is https.
	// +optional
	InsecureSkipTLS bool `json:"insecureSkipTLS,omitempty"`
	// Token of Authorization for http request
	// +optional
	Token string `json:"token,omitempty"`
}

// PlaybookStatus defines the observed state of Playbook.
type PlaybookStatus struct {
	// Statistics statistics of task counts
	Statistics PlaybookStatistics `json:"statistics,omitempty"`
	// Result will record the results detail.
	Result runtime.RawExtension `json:"result,omitempty"`
	// Phase of playbook.
	Phase PlaybookPhase `json:"phase,omitempty"`
	// FailureReason will be set in the event that there is a terminal problem
	// +optional
	FailureReason PlaybookFailedReason `json:"failureReason,omitempty"`
	// FailureMessage will be set in the event that there is a terminal problem
	// +optional
	FailureMessage string `json:"failureMessage,omitempty"`
}

// PlaybookStatistics contains statistics of task counts.
type PlaybookStatistics struct {
	// Total number of tasks
	Total int `json:"total,omitempty"`
	// Number of successful tasks
	Success int `json:"success,omitempty"`
	// Number of failed tasks
	Failed int `json:"failed,omitempty"`
	// Number of ignored tasks
	Ignored int `json:"ignored,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Playbook",type="string",JSONPath=".spec.playbook"
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=".status.taskResult.total"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Playbook is the Schema for the playbooks API.
// Playbook resource executor a playbook.
type Playbook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlaybookSpec   `json:"spec,omitempty"`
	Status PlaybookStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PlaybookList contains a list of Playbook.
type PlaybookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Playbook `json:"items"`
}

// Register Playbook and PlaybookList types with the scheme.
func init() {
	SchemeBuilder.Register(&Playbook{}, &PlaybookList{})
}
