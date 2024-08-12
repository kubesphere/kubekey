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

type TaskPhase string

const (
	TaskPhasePending TaskPhase = "Pending"
	TaskPhaseRunning TaskPhase = "Running"
	TaskPhaseSuccess TaskPhase = "Success"
	TaskPhaseFailed  TaskPhase = "Failed"
	TaskPhaseIgnored TaskPhase = "Ignored"
)

const (
	// TaskAnnotationRole is the absolute dir of task in project.
	TaskAnnotationRole = "kubesphere.io/role"
)

type KubeKeyTaskSpec struct {
	Name        string   `json:"name,omitempty"`
	Hosts       []string `json:"hosts,omitempty"`
	IgnoreError *bool    `json:"ignoreError,omitempty"`
	Retries     int      `json:"retries,omitempty"`

	When       []string             `json:"when,omitempty"`
	FailedWhen []string             `json:"failedWhen,omitempty"`
	Loop       runtime.RawExtension `json:"loop,omitempty"`

	Module   Module `json:"module,omitempty"`
	Register string `json:"register,omitempty"`
}

type Module struct {
	Name string               `json:"name,omitempty"`
	Args runtime.RawExtension `json:"args,omitempty"`
}

type TaskStatus struct {
	RestartCount int              `json:"restartCount,omitempty"`
	Phase        TaskPhase        `json:"phase,omitempty"`
	HostResults  []TaskHostResult `json:"hostResults,omitempty"`
}

type TaskHostResult struct {
	Host   string `json:"host,omitempty"`
	Stdout string `json:"stdout,omitempty"`
	StdErr string `json:"stdErr,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:scope=Namespaced

type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeKeyTaskSpec `json:"spec,omitempty"`
	Status TaskStatus      `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

func (t Task) IsComplete() bool {
	return t.IsSucceed() || t.IsFailed()
}

func (t Task) IsSucceed() bool {
	return t.Status.Phase == TaskPhaseSuccess || t.Status.Phase == TaskPhaseIgnored
}
func (t Task) IsFailed() bool {
	return t.Status.Phase == TaskPhaseFailed && t.Spec.Retries <= t.Status.RestartCount
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
