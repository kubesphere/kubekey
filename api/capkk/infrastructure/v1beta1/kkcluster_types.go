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
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileKKCluster to clean up KK resources associated with KKCluster before
	// removing it from the apiserver.
	KKClusterFinalizer = "kkcluster.infrastructure.cluster.x-k8s.io"
)

const (
	// KKClusterNodeReachedCondition represents the condition type indicating whether the hosts
	// defined in the inventory are reachable.
	KKClusterNodeReachedCondition clusterv1beta1.ConditionType = "NodeReached"
	// KKClusterNodeReachedConditionReasonWaiting indicates that the node reachability check is pending.
	// This check is triggered when the corresponding inventory host's configuration changes.
	KKClusterNodeReachedConditionReasonWaiting = "waiting for node status check"
	// KKClusterNodeReachedConditionReasonUnreached indicates that the node reachability check has failed.
	// This means the node is currently offline or inaccessible.
	KKClusterNodeReachedConditionReasonUnreached = "node is unreachable"

	// KKClusterKKMachineConditionReady represents the condition type indicating whether the associated inventory
	// has been successfully marked as ready.
	KKClusterKKMachineConditionReady clusterv1beta1.ConditionType = "KKClusterReady"
	// KKClusterKKMachineConditionReadyReasonWaiting indicates that the associated inventory is still being synchronized.
	KKClusterKKMachineConditionReadyReasonWaiting = "waiting for kkmachine sync"
	// KKMachineKKMachineConditionReasonSyncing indicates that the associated inventory has been successfully synchronized.
	KKMachineKKMachineConditionReasonSyncing = "syncing for kkmachine"
	// KKMachineKKMachineConditionReasonFailed indicates that the associated inventory synchronization process has failed.
	KKMachineKKMachineConditionReasonFailed = "kkmachine run failed"
)

// HostSelectorPolicy defines the strategy for synchronizing hosts to a specific group.
type HostSelectorPolicy string

const (
	// HostSelectorRandom the host will be selected randomly.
	HostSelectorRandom HostSelectorPolicy = "Random"
	// HostSelectorSequence the host will be selected sequentially.
	HostSelectorSequence HostSelectorPolicy = "Sequence"
)

type KKClusterFailedReason string

const (
	// KKClusterFailedReasonUnknown like cannot get resource from kubernetes.
	KKClusterFailedUnknown KKClusterFailedReason = "unknown"
	// KKClusterFailedReasonInvalidHosts like hosts defined in kkcluster is invalid.
	KKClusterFailedInvalidHosts KKClusterFailedReason = "hosts defined in kkcluster is invalid."
	// KKClusterFailedReasonSyncInventory like failed to sync inventory.
	KKClusterFailedSyncInventory KKClusterFailedReason = "failed to sync inventory"
	// KKClusterFailedReasonSyncCPKKMachine like failed to sync control_plane kkmachine.
	KKClusterFailedSyncCPKKMachine KKClusterFailedReason = "sync control_plane kkmachine failed."
	// KKClusterFailedReasonSyncWorkerKKMachine like failed to sync worker kkmachine.
	KKClusterFailedSyncWorkerKKMachine KKClusterFailedReason = "sync worker kkmachine failed."
)

// ControlPlaneEndpointType defines the type of control plane endpoint used for communication with the cluster.
type ControlPlaneEndpointType string

const (
	// ControlPlaneEndpointTypeDNS indicates the control plane endpoint is a globally resolvable DNS entry.
	// ensuring that the configuration always points to the control plane nodes.
	ControlPlaneEndpointTypeDNS ControlPlaneEndpointType = "dns"
	// ControlPlaneEndpointTypeVIP(DEFAULT) indicates the control plane endpoint is a Virtual IP (VIP).
	// - ARP Mode: Requires the management cluster and worker cluster nodes to be in the same network segment.
	// - BGP Mode: Requires a network environment that supports BGP, with proper configuration in both
	//   the management and worker clusters.
	ControlPlaneEndpointTypeVIP ControlPlaneEndpointType = "vip"
)

type InventoryHostConnector struct {
	// Type to connector the host. de
	Type string `json:"type"`
	// Host address.
	Host string `json:"host"`
	// User is the user name of the host. default is root.
	// +optional
	User string `json:"user,omitempty"`
	// Password is the password of the host.
	// +optional
	Password string `json:"password,omitempty"`
	// PrivateKey is the private key of the host. default is ~/.ssh/id_rsa.
	// +optional
	PrivateKey string `json:"privateKey,omitempty"`
}
type InventoryHost struct {
	// Name of the host.
	Name string `json:"name"`
	// Vars for the host.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars runtime.RawExtension `json:"vars,omitempty"`
	// Connector to connect the host.
	Connector InventoryHostConnector `json:"connector"`
}

// KKClusterSpec defines the desired state of KKCluster.
type KKClusterSpec struct {
	// InventoryHosts contains all hosts of the cluster.
	InventoryHosts []InventoryHost `json:"hosts,omitempty"`
	// which Group defined in Inventory will be checked. there is some default group by system:
	// - all: contains all hosts
	// - ungrouped: contains hosts which do not belong to any groups.
	// if the value is empty, "ungrouped" will be used.
	HostCheckGroup string `json:"hostCheckGroup,omitempty"`
	// tolerate defines if tolerate host check if failed.
	Tolerate bool `json:"tolerate,omitempty"`
	// HostSelectorPolicy defines the strategy for synchronizing hosts to a specific group.
	HostSelectorPolicy HostSelectorPolicy `json:"hostSelectorPolicy,omitempty"`
	// KubeVersion defines the version of kubernetes.
	KubeVersion string `json:"kubeVersion,omitempty"`
	// ControlPlaneEndpointType defines the type of control plane endpoint. such as dns, vip.
	// when use vip, it will deploy kube-vip in each control_plane node. the default value is vip.
	ControlPlaneEndpointType ControlPlaneEndpointType `json:"controlPlaneEndpointType,omitempty"`
}

// KKClusterStatus defines the observed state of KKCluster.
type KKClusterStatus struct {
	// if Ready to create cluster. usage after inventory is ready.
	Ready bool `json:"ready"`

	// FailureReason
	FailureReason KKClusterFailedReason `json:"failureReason,omitempty"`

	FailureMessage string `json:"failureMessage,omitempty"`

	// Conditions defines current service state of the KKCluster.
	// +optional
	Conditions clusterv1beta1.Conditions `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced,categories=cluster-api,shortName=kkc
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1beta1"
// +kubebuilder:printcolumn:name="Cluster",type="string",JSONPath=".metadata.labels.cluster\\.x-k8s\\.io/cluster-name",description="Cluster to which this KKClusters belongs"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.ready",description="Cluster infrastructure is ready for SSH instances"
// +kubebuilder:printcolumn:name="Endpoint",type="string",JSONPath=".spec.controlPlaneEndpoint",description="API Endpoint",priority=1

// KKCluster resource maps a kubernetes cluster, manage and reconcile cluster status.
type KKCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KKClusterSpec   `json:"spec,omitempty"`
	Status KKClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KKClusterList of KKCluster
type KKClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KKCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KKCluster{}, &KKClusterList{})
}

// GetConditions returns the observations of the operational state of the KKCluster resource.
func (k *KKCluster) GetConditions() clusterv1beta1.Conditions {
	return k.Status.Conditions
}

// SetConditions sets the underlying service state of the KKCluster to the predescribed clusterv1beta1.Conditions.
func (k *KKCluster) SetConditions(conditions clusterv1beta1.Conditions) {
	k.Status.Conditions = conditions
}
