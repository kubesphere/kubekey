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
	// ClusterFinalizer allows ReconcileKKCluster to clean up KK resources associated with KKCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "kkcluster.infrastructure.cluster.x-k8s.io"
)

// KKClusterSpec defines the desired state of KKCluster
type KKClusterSpec struct {

	// Nodes represents the information about the nodes available to the cluster
	Nodes Nodes `json:"nodes"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`

	// ControlPlaneLoadBalancer is optional configuration for customizing control plane behavior.
	ControlPlaneLoadBalancer *KKLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// Component is optional configuration for modifying the FTP server
	// that gets the necessary components for the cluster.
	// +optional
	Component *Component `json:"component,omitempty"`

	// Registry represents the cluster image registry used to pull the images.
	// +optional
	Registry Registry `json:"registry,omitempty"`
}

// Nodes represents the information about the nodes available to the cluster
type Nodes struct {
	// Auth is the SSH authentication information of all instance. It is a global auth configuration.
	// +optional
	Auth Auth `json:"auth"`

	// Instances defines all instance contained in kkcluster.
	Instances []InstanceInfo `json:"instances"`
}

// InstanceInfo defines the information about the instance.
type InstanceInfo struct {
	// Name is the host name of the machine.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Address is the IP address of the machine.
	Address string `json:"address,omitempty"`

	// InternalAddress is the internal IP address of the machine.
	InternalAddress string `json:"internalAddress,omitempty"`

	// Roles is the role of the machine.
	// +optional
	Roles []Role `json:"roles,omitempty"`

	// Arch is the architecture of the machine. e.g. "amd64", "arm64".
	// +optional
	Arch string `json:"arch,omitempty"`

	// Auth is the SSH authentication information of this machine. It will override the global auth configuration.
	// +optional
	Auth Auth `json:"auth,omitempty"`
}

// KKLoadBalancerSpec defines the desired state of an KK load balancer.
type KKLoadBalancerSpec struct {
	// The hostname on which the API server is serving.
	Host string `json:"host,omitempty"`
}

// KKClusterStatus defines the observed state of KKCluster
type KKClusterStatus struct {
	// +kubebuilder:default=false
	Ready bool `json:"ready"`
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
// +kubebuilder:resource:path=kkclusters,scope=Namespaced,categories=cluster-api,shortName=kkc
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKClusters belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for SSH instances"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint",description="API Endpoint",priority=1
// +k8s:defaulter-gen=true

// KKCluster is the Schema for the kkclusters API
type KKCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKClusterSpec   `json:"spec,omitempty"`
	Status KKClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KKClusterList contains a list of KKCluster
type KKClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKCluster `json:"items"`
}

// GetConditions returns the observations of the operational state of the KKCluster resource.
func (k *KKCluster) GetConditions() clusterv1.Conditions {
	return k.Status.Conditions
}

// SetConditions sets the underlying service state of the KKCluster to the predescribed clusterv1.Conditions.
func (k *KKCluster) SetConditions(conditions clusterv1.Conditions) {
	k.Status.Conditions = conditions
}

func init() {
	SchemeBuilder.Register(&KKCluster{}, &KKClusterList{})
}
