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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

// InstanceState describes the state of an AWS instance.
type InstanceState string

var (
	// InstanceStatePending is the string representing an instance in a pending state.
	InstanceStatePending = InstanceState("pending")

	InstanceStateBootstrapping = InstanceState("bootstrapping")

	//InstanceStateBootstrapped = InstanceState("bootstrapped")

	// InstanceStateRunning is the string representing an instance in a running state.
	InstanceStateRunning = InstanceState("running")

	// InstanceStateCleaning is the string representing an instance in a cleaning state.
	InstanceStateCleaning = InstanceState("cleaning")

	// InstanceStateCleaned is the string representing an instance in a cleared state.
	InstanceStateCleaned = InstanceState("cleaned")

	// InstanceRunningStates defines the set of states in which an SSH instance is
	// running or going to be running soon.
	InstanceRunningStates = sets.NewString(
		string(InstanceStatePending),
		string(InstanceStateRunning),
	)

	// InstanceKnownStates represents all known EC2 instance states.
	InstanceKnownStates = InstanceRunningStates.Union(
		sets.NewString(
			string(InstanceStateBootstrapping),
			string(InstanceStateCleaning),
			string(InstanceStateCleaned),
		),
	)
)

// KKInstanceSpec defines the desired state of KKInstance
type KKInstanceSpec struct {
	ID string `json:"id"`

	// Name is the host name of the machine.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Address is the IP address of the machine.
	Address string `json:"address"`

	// InternalAddress is the internal IP address of the machine.
	InternalAddress string `json:"internalAddress"`

	// Roles is the role of the machine.
	Roles []Role `json:"roles"`

	// Arch is the architecture of the machine. e.g. "amd64", "arm64".
	Arch string `json:"arch"`

	// Auth is the SSH authentication information of this machine. It will override the global auth configuration.
	Auth Auth `json:"auth,omitempty"`

	// ContainerManager is the container manager config of this machine.
	ContainerManager ContainerManager `json:"containerManager"`

	Bootstrap clusterv1.Bootstrap `json:"bootstrap"`
}

// KKInstanceStatus defines the observed state of KKInstance
type KKInstanceStatus struct {
	// The current state of the instance.
	State InstanceState `json:"instanceState,omitempty"`

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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KKInstance is the Schema for the kkinstances API
type KKInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKInstanceSpec   `json:"spec,omitempty"`
	Status KKInstanceStatus `json:"status,omitempty"`
}

// GetConditions returns the observations of the operational state of the KKMachine resource.
func (k *KKInstance) GetConditions() clusterv1.Conditions {
	return k.Status.Conditions
}

// SetConditions sets the underlying service state of the KKMachine to the predescribed clusterv1.Conditions.
func (k *KKInstance) SetConditions(conditions clusterv1.Conditions) {
	k.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// KKInstanceList contains a list of KKInstance
type KKInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKInstance{}, &KKInstanceList{})
}
