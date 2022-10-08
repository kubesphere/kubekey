/*
Copyright 2022 The KubeSphere Authors.

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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

const (
	// MachineFinalizer allows ReconcileKKMachine to clean up KubeKey resources associated with KKMachine before
	// removing it from the apiserver.
	MachineFinalizer = "kkmachine.infrastructure.cluster.x-k8s.io"
)

// KKMachineSpec defines the desired state of KKMachine
type KKMachineSpec struct {
	// ProviderID is the unique identifier as specified by the kubekey provider.
	ProviderID *string `json:"providerID,omitempty"`

	// InstanceID is the name of the KKInstance.
	InstanceID *string `json:"instanceID,omitempty"`

	// Roles is the role of the machine.
	// +optional
	Roles []Role `json:"roles"`

	// ContainerManager is the container manager config of this machine.
	// +optional
	ContainerManager ContainerManager `json:"containerManager"`

	// Repository is the repository config of this machine.
	// +optional
	Repository *Repository `json:"repository,omitempty"`
}

// KKMachineStatus defines the observed state of KKMachine
type KKMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready"`

	// Addresses contains the KK instance associated addresses.
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// InstanceState is the state of the KK instance for this machine.
	// +optional
	InstanceState *InstanceState `json:"instanceState,omitempty"`

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

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=kkmachines,scope=Namespaced,categories=cluster-api,shortName=kkm
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKMachine belongs"
// +kubebuilder:printcolumn:name="Instance",type="string",JSONPath=".spec.instanceID",description="KKInstance name"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this KKMachine"
// +k8s:defaulter-gen=true

// KKMachine is the Schema for the kkmachines API
type KKMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKMachineSpec   `json:"spec,omitempty"`
	Status KKMachineStatus `json:"status,omitempty"`
}

// GetConditions returns the observations of the operational state of the KKMachine resource.
func (k *KKMachine) GetConditions() clusterv1.Conditions {
	return k.Status.Conditions
}

// SetConditions sets the underlying service state of the KKMachine to the predescribed clusterv1.Conditions.
func (k *KKMachine) SetConditions(conditions clusterv1.Conditions) {
	k.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// KKMachineList contains a list of KKMachine
type KKMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKMachine{}, &KKMachineList{})
}
