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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

type KKMachinePhase string

// KKMachinePhase is KKMachine's current status defined in KKMachine.Status.Phase.
const (
	KKMachinePhaseInitializing KKMachinePhase = "Initializing"
)

// const (
// KUBERNETES the Kubernetes distributions
// KUBERNETES = "kubernetes"
// K3S the K3S distributions
// K3S = "k3s"

// InPlaceUpgradeVersionAnnotation is the annotation that stores the version of the cluster used for in-place upgrade.
// InPlaceUpgradeVersionAnnotation = "kkcluster.infrastructure.cluster.x-k8s.io/in-place-upgrade-version"
// )

// KKMachineSpec defines the desired state of KKMachine
type KKMachineSpec struct {
	// ProviderID is the unique identifier as specified by the kubekey provider.
	ProviderID *string `json:"providerID,omitempty"`
	// Roles is the role of the machine, like `Worker`, `Master`, `ControlPlane`, etc.
	// +optional
	Roles []string `json:"roles"`
}

// KKMachineStatus defines the observed state of KKMachine
type KKMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Phase of KKMachine.
	Phase KKMachinePhase `json:"phase,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a succinct value suitable
	// for machine interpretation.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureReason *errors.MachineStatusError `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// reconciling the Machine and will contain a more verbose string suitable
	// for logging and human consumption.
	//
	// This field should not be set for transitive errors that a controller
	// faces that are expected to be fixed automatically over
	// time (like service outages), but instead indicate that something is
	// fundamentally wrong with the Machine's spec or the configuration of
	// the controller, and that manual intervention is required. Examples
	// of terminal errors would be invalid combinations of settings in the
	// spec, values that are unsupported by the controller, or the
	// responsible controller itself being critically misconfigured.
	//
	// Any transient errors that occur during the reconciliation of Machines
	// can be added as events to the Machine object and/or logged in the
	// controller's output.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Conditions defines current service state of the KKMachine.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kkmachines,scope=Namespaced,categories=cluster-api,shortName=kkm
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKMachine belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this KKMachine"

type KKMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKMachineSpec   `json:"spec,omitempty"`
	Status KKMachineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type KKMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKMachine{}, &KKMachineList{})
}
