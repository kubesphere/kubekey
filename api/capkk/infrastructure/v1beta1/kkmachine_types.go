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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

const (
	// MachineFinalizer allows ReconcileKKMachine to clean up KubeKey resources associated with KKMachine before
	// removing it from the apiserver.
	KKMachineFinalizer = "kkmachine.infrastructure.cluster.x-k8s.io"

	// KKMachineBelongGroupLabel defines which kkmachine belong to.
	KKMachineBelongGroupLabel = "kkmachine.infrastructure.cluster.x-k8s.io/group"

	// AddNodePlaybookAnnotation add node to cluster.
	AddNodePlaybookAnnotation = "playbook.kubekey.kubesphere.io/add-node"
	// DeleteNodePlaybookAnnotation remove node from cluster.
	DeleteNodePlaybookAnnotation = "playbook.kubekey.kubesphere.io/delete-node"
)

type KKMachineFailedReason string

const (
	// KKMachineFailedReasonAddNodeFailed add node failed.
	KKMachineFailedReasonAddNodeFailed KKMachineFailedReason = "add node failed"
	// KKMachineFailedReasonDeleteNodeFailed delete node failed.
	KKMachineFailedReasonDeleteNodeFailed clusterv1beta1.ConditionType = "delete failed failed"
)

// KKMachineSpec defines the desired state of KKMachine.
type KKMachineSpec struct {
	// Roles defines the roles assigned to the Kubernetes cluster node, such as "worker" or "control-plane".
	// A KKMachine created by ControlPlane will automatically have the "control-plane" role.
	// A KKMachine created by MachineDeployment will automatically have the "worker" role.
	// Additional custom roles can also be specified in this field as needed.
	Roles []string `json:"roles,omitempty"`

	// providerID is the identification ID of the machine provided by the provider.
	// This field must match the provider ID as seen on the node object corresponding to this machine.
	// This field is required by higher level consumers of cluster-api. Example use case is cluster autoscaler
	// with cluster-api as provider. Clean-up logic in the autoscaler compares machines to nodes to find out
	// machines at provider which could not get registered as Kubernetes nodes. With cluster-api as a
	// generic out-of-tree provider for autoscaler, this field is required by autoscaler to be
	// able to have a provider view of the list of machines. Another list of nodes is queried from the k8s apiserver
	// and then a comparison is done to find out unregistered machines and are marked for delete.
	// This field will be set by the actuators and consumed by higher level entities like autoscaler that will
	// be interfacing with cluster-api as generic provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`

	// version defines the desired Kubernetes version.
	// This field is meant to be optionally used by bootstrap providers.
	// +optional
	Version *string `json:"version,omitempty"`

	// failureDomain is the failure domain the machine will be created in.
	// Must match a key in the FailureDomains map stored on the cluster object.
	// +optional
	FailureDomain *string `json:"failureDomain,omitempty"`

	// Config for machine. contains cluster version, binary version, etc.
	// + optional
	Config runtime.RawExtension `json:"config,omitempty"`
}

// KKMachineStatus defines the observed state of KKMachine.
type KKMachineStatus struct {
	// Ready is true when the provider resource is ready.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// FailureReason will be set in the event that there is a terminal problem
	// +optional
	FailureReason KKMachineFailedReason `json:"failureReason,omitempty"`

	// FailureMessage will be set in the event that there is a terminal problem
	// +optional
	FailureMessage string `json:"failureMessage,omitempty"`

	// certificatesExpiryDate is the expiry date of the machine certificates.
	// This value is only set for control plane machines.
	// +optional
	CertificatesExpiryDate *metav1.Time `json:"certificatesExpiryDate,omitempty"`

	// Conditions defines current service state of the KKMachine.
	// +optional
	Conditions clusterv1beta1.Conditions `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced,categories=cluster-api,shortName=kkm
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1beta1"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKMachine belongs"
// +kubebuilder:printcolumn:name="ProviderID",type="string",JSONPath=".spec.providerID",description="the providerID for the machine"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Machine ready status"
// +kubebuilder:printcolumn:name="Machine",type="string",JSONPath=".metadata.ownerReferences[?(@.kind==\"Machine\")].name",description="Machine object which owns with this KKMachine"

// KKMachine resource maps a machine instance, manage and reconcile machine status.
type KKMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKMachineSpec   `json:"spec,omitempty"`
	Status KKMachineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KKMachineList of KKMachine
type KKMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKMachine{}, &KKMachineList{})
}
