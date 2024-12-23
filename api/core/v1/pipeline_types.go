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
)

const (
	// BuiltinsProjectAnnotation use builtins project of KubeKey
	BuiltinsProjectAnnotation = "kubekey.kubesphere.io/builtins-project"

	// PipelineCompletedFinalizer will be removed after the Pipeline is completed.
	PipelineCompletedFinalizer = "kubekey.kubesphere.io/pipeline-completed"
)

// PipelinePhase of Pipeline
type PipelinePhase string

const (
	// PipelinePhasePending of Pipeline. Pipeline has created but not deal
	PipelinePhasePending PipelinePhase = "Pending"
	// PipelinePhaseRunning of Pipeline. deal Pipeline.
	PipelinePhaseRunning PipelinePhase = "Running"
	// PipelinePhaseFailed of Pipeline. once Task run failed.
	PipelinePhaseFailed PipelinePhase = "Failed"
	// PipelinePhaseSucceed of Pipeline. all Tasks run success.
	PipelinePhaseSucceeded PipelinePhase = "Succeeded"
)

type PipelineFailedReason string

const (
	// PipelineFailedReasonUnknown is the default failed reason.
	PipelineFailedReasonUnknown PipelineFailedReason = "unknown"
	// PipelineFailedReasonPodFailed pod exec failed.
	PipelineFailedReasonPodFailed PipelineFailedReason = "pod executor failed"
	// PipelineFailedReasonTaskFailed task exec failed.
	PipelineFailedReasonTaskFailed PipelineFailedReason = "task executor failed"
)

// PipelineSpec of pipeline.
type PipelineSpec struct {
	// Project is storage for executable packages
	// +optional
	Project PipelineProject `json:"project,omitempty"`
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
	// If Debug mode is true, It will retain runtime data after a successful execution of Pipeline,
	// which includes task execution status and parameters.
	// +optional
	Debug bool `json:"debug,omitempty"`
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

// PipelineProject respect which playbook store.
type PipelineProject struct {
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

// PipelineStatus of Pipeline
type PipelineStatus struct {
	// TaskResult total related tasks execute result.
	TaskResult PipelineTaskResult `json:"taskResult,omitempty"`
	// Phase of pipeline.
	Phase PipelinePhase `json:"phase,omitempty"`
	// FailureReason will be set in the event that there is a terminal problem
	// +optional
	FailureReason PipelineFailedReason `json:"failureReason,omitempty"`
	// FailureMessage will be set in the event that there is a terminal problem
	// +optional
	FailureMessage string `json:"failureMessage,omitempty"`
	// FailedDetail will record the failed tasks.
	FailedDetail []PipelineFailedDetail `json:"failedDetail,omitempty"`
}

// PipelineTaskResult of Pipeline
type PipelineTaskResult struct {
	// Total number of tasks.
	Total int `json:"total,omitempty"`
	// Success number of tasks.
	Success int `json:"success,omitempty"`
	// Failed number of tasks.
	Failed int `json:"failed,omitempty"`
	// Ignored number of tasks.
	Ignored int `json:"ignored,omitempty"`
}

// PipelineFailedDetail store failed message when pipeline run failed.
type PipelineFailedDetail struct {
	// Task name of failed task.
	Task string `json:"task,omitempty"`
	// failed Hosts Result of failed task.
	Hosts []PipelineFailedDetailHost `json:"hosts,omitempty"`
}

// PipelineFailedDetailHost detail failed message for each host.
type PipelineFailedDetailHost struct {
	// Host name of failed task.
	Host string `json:"host,omitempty"`
	// Stdout of failed task.
	Stdout string `json:"stdout,omitempty"`
	// StdErr of failed task.
	StdErr string `json:"stdErr,omitempty"`
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

// Pipeline resource executor a playbook.
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineList of Pipeline
type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}

// //+kubebuilder:webhook:path=/mutate-kubekey-kubesphere-io-v1beta1-pipeline,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachines,verbs=create;update,versions=v1beta1,name=default.kkmachine.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

// var _ webhook.Defaulter = &Pipeline{}

// // Default implements webhook.Defaulter so a webhook will be registered for the type
// func (k *Pipeline) Default() {

// }
