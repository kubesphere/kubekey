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
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MachineFinalizer is set on PrepareForCreate callback.
	MachineFinalizer = "machine.kubekey.kubesphere.io"

	// MachineControlPlaneLabelName is the label set on machines or related objects that are part of a control plane.
	MachineControlPlaneLabelName = "machine.kubekey.kubesphere.io/control-plane"

	// ExcludeNodeDrainingAnnotation annotation explicitly skips node draining if set.
	ExcludeNodeDrainingAnnotation = "machine.kubekey.kubesphere.io/exclude-node-draining"
)

// MachineSpec defines the desired state of Machine
type MachineSpec struct {
	// ClusterName is the name of the Cluster this object belongs to.
	ClusterName string `json:"clusterName"`

	// Name is the host name of the machine.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Address is the IP address of the machine.
	Address string `json:"address"`

	// InternalAddress is the internal IP address of the machine.
	InternalAddress string `json:"internalAddress"`

	// Arch is the architecture of the machine. e.g. "amd64", "arm64".
	Arch string `json:"arch"`

	// Auth is the SSH authentication information of the machine.
	Auth Auth `json:"auth,omitempty"`

	// ContainerManager is the container manager of the machine.
	ContainerManager ContainerManager `json:"containerManager,omitempty"`

	// NodeDrainTimeout is the total amount of time that the controller will spend on draining a node.
	// The default value is 0, meaning that the node can be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	NodeDrainTimeout *metav1.Duration `json:"nodeDrainTimeout,omitempty"`

	// NodeDeletionTimeout defines how long the controller will attempt to delete the Node that the Machine
	// hosts after the Machine is marked for deletion. A duration of 0 will retry deletion indefinitely.
	// Defaults to 10 seconds.
	// +optional
	NodeDeletionTimeout *metav1.Duration `json:"nodeDeletionTimeout,omitempty"`
}

func (ms *MachineSpec) FillAuth(auth *Auth) error {
	if err := mergo.Merge(&ms.Auth, auth); err != nil {
		return err
	}
	return nil
}

func (ms *MachineSpec) FillContainerManager(containerManager *ContainerManager) error {
	if err := mergo.Merge(&ms.ContainerManager, containerManager); err != nil {
		return err
	}
	return nil
}

// MachineStatus defines the observed state of Machine
type MachineStatus struct {
	// NodeRef will point to the corresponding Node if it exists.
	// +optional
	NodeRef *corev1.ObjectReference `json:"nodeRef,omitempty"`

	// NodeInfo is a set of ids/uuids to uniquely identify the node.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#info
	// +optional
	NodeInfo *corev1.NodeSystemInfo `json:"nodeInfo,omitempty"`

	// LastUpdated identifies when the phase of the Machine last transitioned.
	// +optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

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
	FailureReason *string `json:"failureReason,omitempty"`

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

	// Phase represents the current phase of machine actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`

	// SSHReady is the state of the infrastructure provider.
	// +optional
	SSHReady bool `json:"sshReady,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the Machine.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`
}

// SetTypedPhase sets the Phase field to the string representation of MachinePhase.
func (m *MachineStatus) SetTypedPhase(p MachinePhase) {
	m.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed MachinePhase representation as described in `machine_phase_types.go`.
func (m *MachineStatus) GetTypedPhase() MachinePhase {
	switch phase := MachinePhase(m.Phase); phase {
	case
		MachinePhasePending,
		MachinePhaseRunning,
		MachinePhaseDeleting,
		MachinePhaseDeleted,
		MachinePhaseFailed:
		return phase
	default:
		return MachinePhaseUnknown
	}
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:storageversion

// Machine is the Schema for the machines API
type Machine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineSpec   `json:"spec,omitempty"`
	Status MachineStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (m *Machine) GetConditions() Conditions {
	return m.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (m *Machine) SetConditions(conditions Conditions) {
	m.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// MachineList contains a list of Machine
type MachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Machine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Machine{}, &MachineList{})
}
