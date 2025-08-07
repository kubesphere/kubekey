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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TaskPhase of Task
type TaskPhase string

const (
	// TaskPhasePending of Task. Task has created but not deal
	TaskPhasePending TaskPhase = "Pending"
	// TaskPhaseRunning of Task. deal Task
	TaskPhaseRunning TaskPhase = "Running"
	// TaskPhaseSuccess of Task. Module of Task run success in each hosts.
	TaskPhaseSuccess TaskPhase = "Success"
	// TaskPhaseFailed of Task. once host run failed.
	TaskPhaseFailed TaskPhase = "Failed"
	// TaskPhaseIgnored of Task. once host run failed and set ignore_errors.
	TaskPhaseIgnored TaskPhase = "Ignored"
)

const (
	// TaskAnnotationRelativePath is the relative dir of task in project.
	TaskAnnotationRelativePath = "kubesphere.io/rel-path"
)

// TaskSpec of Task
type TaskSpec struct {
	Name        string   `json:"name,omitempty"`
	Hosts       []string `json:"hosts,omitempty"`
	DelegateTo  string   `yaml:"delegate_to,omitempty"`
	IgnoreError *bool    `json:"ignoreError,omitempty"`
	Retries     int      `json:"retries,omitempty"`

	When       []string             `json:"when,omitempty"`
	FailedWhen []string             `json:"failedWhen,omitempty"`
	Loop       runtime.RawExtension `json:"loop,omitempty"`

	Module       Module `json:"module,omitempty"`
	Register     string `json:"register,omitempty"`
	RegisterType string `json:"register_type,omitempty"`
}

// Module of Task
type Module struct {
	Name string               `json:"name,omitempty"`
	Args runtime.RawExtension `json:"args,omitempty"`
}

// TaskStatus of Task
type TaskStatus struct {
	RestartCount int              `json:"restartCount,omitempty"`
	Phase        TaskPhase        `json:"phase,omitempty"`
	HostResults  []TaskHostResult `json:"hostResults,omitempty"`
}

// TaskHostResult each host result for task
type TaskHostResult struct {
	Host   string `json:"host,omitempty"`
	Stdout string `json:"stdout,omitempty"`
	StdErr string `json:"stdErr,omitempty"`
	Error  string `json:"error,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Namespaced

// Task of playbook
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TaskList for Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

// IsComplete if Task IsSucceed or IsFailed
func (t Task) IsComplete() bool {
	return t.IsSucceed() || t.IsFailed()
}

// IsSucceed if Task.Status.Phase TaskPhaseSuccess or TaskPhaseIgnored
func (t Task) IsSucceed() bool {
	return t.Status.Phase == TaskPhaseSuccess || t.Status.Phase == TaskPhaseIgnored
}

// IsFailed Task.Status.Phase is failed when reach the retries
func (t Task) IsFailed() bool {
	return t.Status.Phase == TaskPhaseFailed && t.Spec.Retries <= t.Status.RestartCount
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
