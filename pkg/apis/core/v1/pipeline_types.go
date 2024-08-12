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

type PipelinePhase string

const (
	PipelinePhasePending PipelinePhase = "Pending"
	PipelinePhaseRunning PipelinePhase = "Running"
	PipelinePhaseFailed  PipelinePhase = "Failed"
	PipelinePhaseSucceed PipelinePhase = "Succeed"
)

const (
	// BuiltinsProjectAnnotation use builtins project of KubeKey
	BuiltinsProjectAnnotation = "kubekey.kubesphere.io/builtins-project"
)

type PipelineSpec struct {
	// Project is storage for executable packages
	// +optional
	Project PipelineProject `json:"project,omitempty"`
	// Playbook which to execute.
	Playbook string `json:"playbook"`
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

// PipelineJobSpec set the spec of the job that allows configuration
// Each time the pipeline is executed, usually only one pod is created and will not be re-executed after failure.
type PipelineJobSpec struct {
	// when Schedule is not empty, pipeline will create CornJob, otherwise pipeline will create Job.
	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// The number of successful finished jobs to retain. Value must be non-negative integer.
	// Defaults to 3.
	// +optional
	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty"`

	// The number of failed finished jobs to retain. Value must be non-negative integer.
	// Defaults to 1.
	// +optional
	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty"`

	// suspend specifies whether the Job controller should create Pods or not. If
	// a Job is created with suspend set to true, no Pods are created by the Job
	// controller. If a Job is suspended after creation (i.e. the flag goes from
	// false to true), the Job controller will delete all active Pods associated
	// with this Job. Users must design their workload to gracefully handle this.
	// Suspending a Job will reset the StartTime field of the Job, effectively
	// resetting the ActiveDeadlineSeconds timer too. Defaults to false.
	//
	// +optional
	Suspend *bool `json:"suspend,omitempty"`

	// Specifies the duration in seconds relative to the startTime that the job
	// may be continuously active before the system tries to terminate it; value
	// must be positive integer. If a Job is suspended (at creation or through an
	// update), this timer will effectively be stopped and reset when the Job is
	// resumed again.
	// +optional
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`

	// ttlSecondsAfterFinished limits the lifetime of a Job that has finished
	// execution (either Complete or Failed). If this field is set,
	// ttlSecondsAfterFinished after the Job finishes, it is eligible to be
	// automatically deleted. When the Job is being deleted, its lifecycle
	// guarantees (e.g. finalizers) will be honored. If this field is unset,
	// the Job won't be automatically deleted. If this field is set to zero,
	// the Job becomes eligible to be deleted immediately after it finishes.
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`

	// Volumes in job pod.
	// +optional
	Volumes []corev1.Volume `json:"workVolume,omitempty"`

	// VolumeMounts in job pod.
	// +optional
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

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

type PipelineStatus struct {
	// TaskResult total related tasks execute result.
	TaskResult PipelineTaskResult `json:"taskResult,omitempty"`
	// Phase of pipeline.
	Phase PipelinePhase `json:"phase,omitempty"`
	// failed Reason of pipeline.
	Reason string `json:"reason,omitempty"`
	// FailedDetail will record the failed tasks.
	FailedDetail []PipelineFailedDetail `json:"failedDetail,omitempty"`
}

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

type PipelineFailedDetail struct {
	// Task name of failed task.
	Task string `json:"task,omitempty"`
	// failed Hosts Result of failed task.
	Hosts []PipelineFailedDetailHost `json:"hosts,omitempty"`
}

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

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
