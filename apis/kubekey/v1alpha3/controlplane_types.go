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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ControlPlaneSpec defines the desired state of ControlPlane
type ControlPlaneSpec struct {
	// Replicas is the number of desired replicas.
	Replicas *int32 `json:"replicas,omitempty"`

	// ClusterName is the name of the Cluster
	ClusterName string `json:"clusterName"`

	// Version defines the desired Kubernetes version.
	Version string `json:"version"`

	// Machines defines machines contained in the control plane
	Machines KubeadmControlPlaneMachines `json:"machines"`

	// KubeadmConfigTemplateSpec is a KubeadmConfigTemplateSpec
	// to use for initializing and joining machines to the control plane.
	KubeadmConfigTemplateSpec KubeadmConfigTemplateSpec `json:"kubeadmConfigTemplateSpec"`
}

// KubeadmControlPlaneMachines contains information about how machines
// in a KubeadmControlPlane object.
type KubeadmControlPlaneMachines struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta ObjectMeta `json:"metadata,omitempty"`

	// InfrastructureRef is a required reference to a custom resource
	// offered by an infrastructure provider.
	InfrastructureRef corev1.ObjectReference `json:"infrastructureRef"`
}

// ControlPlaneStatus defines the observed state of ControlPlane
type ControlPlaneStatus struct {
	// Selector is the label selector in string format to avoid introspection
	// by clients, and is used to provide the CRD-based integration for the
	// scale subresource and additional integrations for things like kubectl
	// describe. The string will be in the same format as the query-param syntax.
	// More info about label selectors: http://kubernetes.io/docs/user-guide/labels#label-selectors
	// +optional
	Selector string `json:"selector,omitempty"`

	// Total number of non-terminated machines targeted by this control plane
	// (their labels match the selector).
	// +optional
	Replicas int32 `json:"replicas"`

	// Version represents the minimum Kubernetes version for the control plane machines
	// in the cluster.
	// +optional
	Version *string `json:"version,omitempty"`

	// Total number of non-terminated machines targeted by this control plane
	// that have the desired template spec.
	// +optional
	UpdatedReplicas int32 `json:"updatedReplicas"`

	// Total number of fully running and ready control plane machines.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas"`

	// Initialized denotes whether or not the control plane has the
	// uploaded kubeadm-config configmap.
	// +optional
	Initialized bool `json:"initialized"`

	// Ready denotes that the KubeadmControlPlane API Server is ready to
	// receive requests.
	// +optional
	Ready bool `json:"ready"`

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// ErrorMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the KubeadmControlPlane.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:storageversion

// ControlPlane is the Schema for the controlplanes API
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (c *ControlPlane) GetConditions() Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *ControlPlane) SetConditions(conditions Conditions) {
	c.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// ControlPlaneList contains a list of ControlPlane
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
