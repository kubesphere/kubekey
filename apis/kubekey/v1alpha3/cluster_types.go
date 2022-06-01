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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// WatchLabel is a label othat can be applied to any Cluster API object.
	//
	// Controllers which allow for selective reconciliation may check this label and proceed
	// with reconciliation of the object only if this label and a configured value is present.
	WatchLabel = "kubekey.kubesphere.io/watch-filter"

	// PausedAnnotation is an annotation that can be applied to any Cluster API
	// object to prevent a controller from processing a resource.
	//
	// Controllers working with Cluster API objects must check the existence of this annotation
	// on the reconciled object.
	PausedAnnotation = "kubekey.kubesphere.io/paused"

	// MachineSkipRemediationAnnotation is the annotation used to mark the machines that should not be considered for remediation by MachineHealthCheck reconciler.
	MachineSkipRemediationAnnotation = "kubekey.kubesphere.io/skip-remediation"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Paused can be used to prevent controllers from processing the Cluster and all its associated objects.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint ControlPlaneEndpoint `json:"controlPlaneEndpoint,omitempty"`

	// Network defines the network configuration for the cluster.
	// +optional
	Network *Network `json:"network,omitempty"`

	// RegistryConfig defines the registry configuration for the cluster.
	// +optional
	RegistryConfig *RegistryConfig `json:"registryConfig,omitempty"`

	// ETCDRef offered by an etcd custom resource.
	// +optional
	ETCDRef *corev1.ObjectReference `json:"etcdRef"`

	// ControlPlaneRef is an optional reference to a provider-specific resource that holds
	// the details for provisioning the Control Plane for a Cluster.
	// +optional
	ControlPlaneRef *corev1.ObjectReference `json:"controlPlaneRef"`

	// WorkerRefs offered by a worker custom resource.
	// +optional
	WorkerRefs []*corev1.ObjectReference `json:"workerRefs"`

	// RegistryRef us offered by a registry custom resource.
	// +optional
	RegistryRef *corev1.ObjectReference `json:"registryRef"`

	// AddonRef offered by an addon custom resource.
	AddonRef []*corev1.ObjectReference `json:"customRef"`
}

type ControlPlaneEndpoint struct {
	// InternalLoadbalancer defines the type of loadbalancer to use for the control plane.
	// "haproxy" is the only supported value.
	InternalLoadbalancer *string `json:"internalLoadbalancer,omitempty"`

	// Domain defines the control plane domain.
	// Default is "lb.kubesphere.local"
	Domain string `json:"domain,omitempty"`

	// Address defines the control plane api endpoint address.
	Address string `json:"address,omitempty"`

	// Port defines the control plane api endpoint port.
	Port int `json:"port,omitempty"`
}

type Network struct {
	// APIServerPort specifies the port the API Server should bind to.
	// Defaults to 6443.
	// +optional
	APIServerPort *int32 `json:"apiServerPort,omitempty"`

	// The network ranges from which service VIPs are allocated.
	// +optional
	Services *NetworkRanges `json:"services,omitempty"`

	// The network ranges from which Pod networks are allocated.
	// +optional
	Pods *NetworkRanges `json:"pods,omitempty"`

	// Domain name for services.
	// +optional
	ServiceDomain string `json:"serviceDomain,omitempty"`
}

// NetworkRanges represents ranges of network addresses.
type NetworkRanges struct {
	CIDRBlocks []string `json:"cidrBlocks"`
}

func (n NetworkRanges) String() string {
	if len(n.CIDRBlocks) == 0 {
		return ""
	}
	return strings.Join(n.CIDRBlocks, ",")
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// FailureReason indicates that there is a fatal problem reconciling the
	// state, and will be set to a token value suitable for
	// programmatic interpretation.
	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// FailureMessage indicates that there is a fatal problem reconciling the
	// state, and will be set to a descriptive error message.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`

	// InfrastructureReady is the state of the infrastructure provider.
	// +optional
	InfrastructureReady bool `json:"infrastructureReady"`

	// RegistryReady defines if the registry is ready.
	// +optional
	RegistryReady *bool `json:"registryReady, omitempty"`

	// ETCDReady defines if the etcd cluster is ready.
	// +optional
	ETCDReady *bool `json:"etcdReady,omitempty"`

	// ControlPlaneReady defines if the control plane is ready.
	// +optional
	ControlPlaneReady *bool `json:"controlPlaneReady"`

	// WorkerReady defines if the worker is ready.
	// +optional
	WorkerReady *bool `json:"workerReady"`

	// Conditions defines current service state of the cluster.
	// +optional
	Conditions Conditions `json:"conditions,omitempty"`

	// ObservedGeneration is the latest generation observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:storageversion

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// GetConditions returns the set of conditions for this object.
func (c *Cluster) GetConditions() Conditions {
	return c.Status.Conditions
}

// SetConditions sets the conditions on this object.
func (c *Cluster) SetConditions(conditions Conditions) {
	c.Status.Conditions = conditions
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
