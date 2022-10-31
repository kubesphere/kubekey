/*
Copyright 2022.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
)

// RolloutStrategyType defines the rollout strategies for a K3sControlPlane.
type RolloutStrategyType string

const (
	// RollingUpdateStrategyType replaces the old control planes by new one using rolling update
	// i.e. gradually scale up or down the old control planes and scale up or down the new one.
	RollingUpdateStrategyType RolloutStrategyType = "RollingUpdate"
)

const (
	// K3sControlPlaneFinalizer is the finalizer applied to K3sControlPlane resources
	// by its managing controller.
	K3sControlPlaneFinalizer = "k3s.controlplane.cluster.x-k8s.io"

	// SkipCoreDNSAnnotation annotation explicitly skips reconciling CoreDNS if set.
	SkipCoreDNSAnnotation = "controlplane.cluster.x-k8s.io/skip-coredns"

	// SkipKubeProxyAnnotation annotation explicitly skips reconciling kube-proxy if set.
	SkipKubeProxyAnnotation = "controlplane.cluster.x-k8s.io/skip-kube-proxy"

	// K3sServerConfigurationAnnotation is a machine annotation that stores the json-marshalled string of K3SCP ClusterConfiguration.
	// This annotation is used to detect any changes in ClusterConfiguration and trigger machine rollout in K3SCP.
	K3sServerConfigurationAnnotation = "controlplane.cluster.x-k8s.io/k3s-server-configuration"
)

// K3sControlPlaneSpec defines the desired state of K3sControlPlane
type K3sControlPlaneSpec struct {
	// Number of desired machines. Defaults to 1. When stacked etcd is used only
	// odd numbers are permitted, as per [etcd best practice](https://etcd.io/docs/v3.3.12/faq/#why-an-odd-number-of-cluster-members).
	// This is a pointer to distinguish between explicit zero and not specified.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Version defines the desired K3s version.
	Version string `json:"version"`

	// MachineTemplate contains information about how machines
	// should be shaped when creating or updating a control plane.
	MachineTemplate K3sControlPlaneMachineTemplate `json:"machineTemplate"`

	// K3sConfigSpec is a K3sConfigSpec
	// to use for initializing and joining machines to the control plane.
	// +optional
	K3sConfigSpec infrabootstrapv1.K3sConfigSpec `json:"k3sConfigSpec,omitempty"`

	// RolloutAfter is a field to indicate a rollout should be performed
	// after the specified time even if no changes have been made to the
	// K3sControlPlane.
	//
	// +optional
	RolloutAfter *metav1.Time `json:"rolloutAfter,omitempty"`

	// The RolloutStrategy to use to replace control plane machines with
	// new ones.
	// +optional
	// +kubebuilder:default={type: "RollingUpdate", rollingUpdate: {maxSurge: 1}}
	RolloutStrategy *RolloutStrategy `json:"rolloutStrategy,omitempty"`
}

// K3sControlPlaneMachineTemplate defines the template for Machines
// in a K3sControlPlane object.
type K3sControlPlaneMachineTemplate struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`

	// InfrastructureRef is a required reference to a custom resource
	// offered by an infrastructure provider.
	InfrastructureRef corev1.ObjectReference `json:"infrastructureRef"`

	// NodeDrainTimeout is the total amount of time that the controller will spend on draining a controlplane node
	// The default value is 0, meaning that the node can be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	NodeDrainTimeout *metav1.Duration `json:"nodeDrainTimeout,omitempty"`

	// NodeDeletionTimeout defines how long the machine controller will attempt to delete the Node that the Machine
	// hosts after the Machine is marked for deletion. A duration of 0 will retry deletion indefinitely.
	// If no value is provided, the default value for this property of the Machine resource will be used.
	// +optional
	NodeDeletionTimeout *metav1.Duration `json:"nodeDeletionTimeout,omitempty"`
}

// RolloutStrategy describes how to replace existing machines
// with new ones.
type RolloutStrategy struct {
	// Type of rollout. Currently the only supported strategy is
	// "RollingUpdate".
	// Default is RollingUpdate.
	// +optional
	Type RolloutStrategyType `json:"type,omitempty"`

	// Rolling update config params. Present only if
	// RolloutStrategyType = RollingUpdate.
	// +optional
	RollingUpdate *RollingUpdate `json:"rollingUpdate,omitempty"`
}

// RollingUpdate is used to control the desired behavior of rolling update.
type RollingUpdate struct {
	// The maximum number of control planes that can be scheduled above or under the
	// desired number of control planes.
	// Value can be an absolute number 1 or 0.
	// Defaults to 1.
	// Example: when this is set to 1, the control plane can be scaled
	// up immediately when the rolling update starts.
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// K3sControlPlaneStatus defines the observed state of K3sControlPlane
type K3sControlPlaneStatus struct {
	// Selector is the label selector in string format to avoid introspection
	// by clients, and is used to provide the CRD-based integration for the
	// scale subresource and additional integrations for things like kubectl
	// describe.. The string will be in the same format as the query-param syntax.
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

	// Total number of unavailable machines targeted by this control plane.
	// This is the total number of machines that are still required for
	// the deployment to have 100% available capacity. They may either
	// be machines that are running but not yet ready or machines
	// that still have not been created.
	// +optional
	UnavailableReplicas int32 `json:"unavailableReplicas"`

	// Initialized denotes whether or not the control plane has the
	// uploaded kubeadm-config configmap.
	// +optional
	Initialized bool `json:"initialized"`

	// Ready denotes that the K3sControlPlane API Server is ready to
	// receive requests.
	// +optional
	Ready bool `json:"ready"`

	// FailureReason indicates that there is a terminal problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	FailureReason errors.KubeadmControlPlaneStatusError `json:"failureReason,omitempty"`

	// ErrorMessage indicates that there is a terminal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the K3sControlPlane.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=k3scontrolplanes,shortName=k3scp,scope=Namespaced,categories=cluster-api
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// K3sControlPlane is the Schema for the k3scontrolplanes API
type K3sControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K3sControlPlaneSpec   `json:"spec,omitempty"`
	Status K3sControlPlaneStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (in *K3sControlPlane) GetConditions() clusterv1.Conditions {
	return in.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (in *K3sControlPlane) SetConditions(conditions clusterv1.Conditions) {
	in.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// K3sControlPlaneList contains a list of K3sControlPlane
type K3sControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3sControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3sControlPlane{}, &K3sControlPlaneList{})
}
