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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"
)

type KKClusterPhase string

// KKClusterPhase is KKCluster's current status defined in KKCluster.Status.Phase.
const (
	KKClusterPhasePending KKClusterPhase = "Pending"
	KKClusterPhaseSucceed KKClusterPhase = "Succeed"
	KKClusterPhaseRunning KKClusterPhase = "Running"
	KKClusterPhaseFailed  KKClusterPhase = "Failed"
)

const (
	// HostReadyCondition will execute inventory_list host 检查是否连通，将 host 创建为 machine (只检查选好的 host)
	HostReadyCondition clusterv1.ConditionType = "HostReadyCondition"
	// EtcdReadyCondition will execute install etcd
	EtcdReadyCondition clusterv1.ConditionType = "EtcdReadyCondition"
	// ClusterBinaryReadyCondition will execute create k8s cluster
	ClusterBinaryReadyCondition clusterv1.ConditionType = "ClusterBinaryReadyCondition"
	// BootstrapReadyCondition will execute install k8s cluster
	BootstrapReadyCondition clusterv1.ConditionType = "BootstrapReadyCondition"
	// ClusterReadyCondition will execute postcheck
	ClusterReadyCondition clusterv1.ConditionType = "ClusterReadyCondition"
)

const (
	// ClusterFinalizer allows ReconcileKKCluster to clean up KK resources associated with KKCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "kkcluster.infrastructure.cluster.x-k8s.io"

	// KUBERNETES the Kubernetes distributions
	// KUBERNETES = "kubernetes"
	// K3S the K3S distributions
	// K3S = "k3s"

	// InPlaceUpgradeVersionAnnotation is the annotation that stores the version of the cluster used for in-place upgrade.
	// InPlaceUpgradeVersionAnnotation = "kkcluster.infrastructure.cluster.x-k8s.io/in-place-upgrade-version"
)

// KKClusterSpec defines the desired state of KKCluster
type KKClusterSpec struct {
	// Distribution represents the Kubernetes distribution type of the cluster.
	Distribution string `json:"distribution,omitempty"`

	// InventoryRef is the node configuration for playbook
	// +optional
	InventoryRef *corev1.ObjectReference `json:"inventoryRef,omitempty"`

	// ControlPlaneLoadBalancer is optional configuration for customizing control plane behavior.
	// +optional
	ControlPlaneLoadBalancer *KKLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
}

// KKClusterStatus defines the observed state of KKCluster
type KKClusterStatus struct {
	// +kubebuilder:default=false
	Ready bool `json:"ready"`

	// Phase of KKCluster.
	Phase KKClusterPhase `json:"phase,omitempty"`

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

// KKLoadBalancerSpec defines the desired state of an KK load balancer.
type KKLoadBalancerSpec struct {
	// The hostname on which the API server is serving.
	Host string `json:"host,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=kkcluster,scope=Namespaced,categories=cluster-api,shortName=kkc
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKClusters belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for SSH instances"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint",description="API Endpoint",priority=1

type KKCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKClusterSpec   `json:"spec,omitempty"`
	Status KKClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

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
