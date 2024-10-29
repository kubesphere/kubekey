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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/errors"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
)

// KKClusterPhase of KKCluster
type KKClusterPhase string

// This const defines current Phase of KKCluster.
const (
	KKClusterPhasePending  KKClusterPhase = "Pending"
	KKClusterPhaseSucceed  KKClusterPhase = "Succeed"
	KKClusterPhaseRunning  KKClusterPhase = "Running"
	KKClusterPhaseFailed   KKClusterPhase = "Failed"
	KKClusterPhaseDeleting KKClusterPhase = "Deleting"
)

// NodeSelectorMode defines selector function during select cluster nodes.
type NodeSelectorMode string

// This const defines various NodeSelectorMode of KKCluster.
const (
	RandomNodeSelectorMode   NodeSelectorMode = "Random"
	SequenceNodeSelectorMode NodeSelectorMode = "Sequence"
	// ResponseTimeNodeSelectorMode NodeSelectorMode = "ResponseTime"
)

// This const defines various Condition, Reason and Message of KKCluster Conditions.
const (
	// HostsReadyCondition will check if hosts are connected firstly, then select control-plane nodes and worker nodes.
	HostsReadyCondition clusterv1.ConditionType = "HostsReadyCondition"
	// WaitingCheckHostReadyReason is one `Reason` of HostsReadyCondition.
	WaitingCheckHostReadyReason string = "WaitingForCheckHosts"
	// WaitingCheckHostReadyMessage is a specification `Message` of WaitingCheckHostReadyReason.
	WaitingCheckHostReadyMessage string = "Waiting for check if all of the hosts are connected."
	// HostsNotReadyReason is one `Reason` of HostsReadyCondition.
	HostsNotReadyReason string = "HostsNotReady"
	// HostsSelectFailedMessage is a specification `Message` of HostsNotReadyReason.
	HostsSelectFailedMessage string = "Not enough connected hosts to complete host selection."
	// WaitingHostsSelectReason is one `Reason` of HostsReadyCondition.
	WaitingHostsSelectReason string = "WaitingForHostsSelect"
	// WaitingHostsSelectMessage is a specification `Message` of HostsNotReadyReason.
	WaitingHostsSelectMessage string = "Waiting for select kube-control-plane and worker nodes."
	// HostsReadyReason is one `Reason` of HostsReadyCondition.
	HostsReadyReason string = "HostsReady"
	// HostsReadyMessage is a specification `Message` of HostsNotReadyReason.
	HostsReadyMessage string = "All hosts are connected."

	// PreparationReadyCondition will check which artifacts need to be installed, also initialize the os system.
	PreparationReadyCondition clusterv1.ConditionType = "PreCheckReadyCondition"
	// WaitingPreparationReason is one `Reason` of PreparationReadyCondition.
	WaitingPreparationReason string = "WaitingForPreparation"
	// WaitingPreparationMessage is a specification `Message` of PreparationReadyCondition.
	WaitingPreparationMessage string = "Waiting for pre-check and pre-install artifacts and initialize os system"
	// PreparationNotReadyReason is one `Reason` of PreparationReadyCondition.
	PreparationNotReadyReason string = "PreparationNotReady"
	// PreparationReadyReason is one `Reason` of PreparationReadyCondition.
	PreparationReadyReason string = "PreparationReady"
	// PreparationReadyMessage is a specification `Message` of PreparationReadyCondition.
	PreparationReadyMessage string = "Both artifacts pre-install and os initialization are ready."

	// EtcdReadyCondition will install etcd into etcd group (binary install only currently).
	EtcdReadyCondition clusterv1.ConditionType = "EtcdReadyCondition"
	// WaitingInstallEtcdReason is one `Reason` of EtcdReadyCondition.
	WaitingInstallEtcdReason string = "WaitingForInstallEtcd"
	// WaitingInstallEtcdMessage is a specification `Message` of EtcdReadyCondition.
	WaitingInstallEtcdMessage string = "Waiting for install ETCD binary service"
	// EtcdNotReadyReason is one `Reason` of EtcdReadyCondition.
	EtcdNotReadyReason string = "ETCDNotReady"
	// EtcdReadyReason is one `Reason` of EtcdReadyCondition.
	EtcdReadyReason string = "EtcdReady"
	// EtcdReadyMessage is a specification `Message` of EtcdReadyCondition.
	EtcdReadyMessage string = "Etcd successfully installed."

	// BinaryInstallCondition will install cluster binary tools.
	BinaryInstallCondition clusterv1.ConditionType = "BinaryInstallCondition"
	// WaitingInstallClusterBinaryReason is one `Reason` of BinaryInstallCondition.
	WaitingInstallClusterBinaryReason string = "WaitingForInstallClusterBinary"
	// WaitingInstallClusterBinaryMessage is a specification `Message` of BinaryInstallCondition.
	WaitingInstallClusterBinaryMessage string = "Waiting for install cluster binary tools, e.g. kubeadm and kubelet, etc."
	// BinaryNotReadyReason is one `Reason` of BinaryInstallCondition.
	BinaryNotReadyReason string = "ClusterBinaryNotReady"
	// BinaryReadyReason is one `Reason` of BinaryInstallCondition.
	BinaryReadyReason string = "ClusterBinaryReady"
	// BinaryReadyMessage is a specification `Message` of BinaryInstallCondition.
	BinaryReadyMessage string = "Cluster binary successfully installed"

	// BootstrapReadyCondition will execute `kubeadm join` & `kubeadm init` command.
	BootstrapReadyCondition clusterv1.ConditionType = "BootstrapReadyCondition"
	// WaitingCheckBootstrapReadyReason is one `Reason` of BootstrapReadyCondition.
	WaitingCheckBootstrapReadyReason string = "WaitingForBootstrapReady"
	// WaitingCheckBootstrapReadyMessage is a specification `Message` of BootstrapReadyCondition.
	WaitingCheckBootstrapReadyMessage string = "Waiting for the initial bootstrap to complete. Adding control plane and worker nodes to the cluster."
	// BootstrapNotReadyReason is one `Reason` of BootstrapReadyCondition.
	BootstrapNotReadyReason string = "CheckBootstrapNotReady"
	// BootstrapReadyReason is one `Reason` of BootstrapReadyCondition.
	BootstrapReadyReason string = "CheckBootstrapReady"
	// BootstrapReadyMessage is a specification `Message` of BootstrapReadyCondition.
	BootstrapReadyMessage string = "Bootstrap is ready."

	// ClusterReadyCondition will check if cluster is ready.
	ClusterReadyCondition clusterv1.ConditionType = "ClusterReadyCondition"
	// WaitingCheckClusterReadyReason is one `Reason` of ClusterReadyCondition.
	WaitingCheckClusterReadyReason string = "WaitingForClusterReady"
	// WaitingCheckClusterReadyMessage is a specification `Message` of ClusterReadyCondition.
	WaitingCheckClusterReadyMessage string = "Waiting for initial bootstrap to ready, add control-plane and worker nodes into cluster."
	// ClusterNotReadyReason is one `Reason` of ClusterReadyCondition.
	ClusterNotReadyReason string = "ClusterNotReady"
	// ClusterReadyReason is one `Reason` of ClusterReadyCondition.
	ClusterReadyReason string = "ClusterReady"
	// ClusterReadyMessage is a specification `Message` of ClusterReadyCondition.
	ClusterReadyMessage string = "Cluster is ready."

	// ClusterDeletingCondition will delete the cluster.
	ClusterDeletingCondition clusterv1.ConditionType = "ClusterDeletingCondition"
	// WaitingClusterDeletingReason is one `Reason` of ClusterDeletingCondition.
	WaitingClusterDeletingReason string = "WaitingForClusterDeleting"
	// WaitingClusterDeletingMessage is a specification `Message` of ClusterDeletingCondition.
	WaitingClusterDeletingMessage string = "Waiting for cluster deletion"
	// ClusterDeletingSucceedReason is one `Reason` of ClusterDeletingCondition.
	ClusterDeletingSucceedReason string = "ClusterDeletingSucceeded"
	// ClusterDeletingFailedReason is one `Reason` of ClusterDeletingCondition.
	ClusterDeletingFailedReason string = "ClusterDeletingFailed"
	// ClusterDeletingSucceedMessage is a specification `Message` of ClusterDeletingCondition.
	ClusterDeletingSucceedMessage string = "Cluster deletion succeeded"
)

// This const defines various annotation of KKCluster.
const (
	// ClusterGroupNameAnnotation defines cluster group name of the kubernetes cluster.
	ClusterGroupNameAnnotation string = "capkk.kubekey.kubesphere.io/cluster-group-name"
	// ControlPlaneGroupNameAnnotation defines control plane group name of the kubernetes cluster.
	ControlPlaneGroupNameAnnotation string = "capkk.kubekey.kubesphere.io/control-plane-group-name"
	// WorkerGroupNameAnnotation defines worker group name of the kubernetes cluster.
	WorkerGroupNameAnnotation string = "capkk.kubekey.kubesphere.io/worker-group-name"
	// EtcdGroupNameAnnotation defines etcd group name of the kubernetes cluster.
	EtcdGroupNameAnnotation string = "capkk.kubekey.kubesphere.io/etcd-group-name"
	// RegistryGroupNameAnnotation defines registry group name of the kubernetes cluster.
	RegistryGroupNameAnnotation string = "capkk.kubekey.kubesphere.io/registry-group-name"
	// KCPSecretsRefreshAnnotation defines if kcp secrets need to refresh
	KCPSecretsRefreshAnnotation string = "capkk.kubekey.kubesphere.io/kcp-secrets-refresh" //nolint:gosec
)

// This const defines default node select strategy way and annotations values of KKCluster.
const (
	// DefaultNodeSelectorMode is the default node selector mode for the cluster.
	DefaultNodeSelectorMode = "Random"
	// DefaultClusterGroupName is the default cluster group name.
	DefaultClusterGroupName = "k8s_cluster"
	// DefaultControlPlaneGroupName is the default control plane group name.
	DefaultControlPlaneGroupName = "kube_control_plane"
	// DefaultWorkerGroupName is the default worker group name.
	DefaultWorkerGroupName = "kube_worker"
	// DefaultEtcdGroupName is the default etcd group name.
	DefaultEtcdGroupName = "etcd"
	// DefaultRegistryGroupName is the default registry group name.
	DefaultRegistryGroupName = "registry"
)

const (
	// ClusterFinalizer allows ReconcileKKCluster to clean up KK resources associated with KKCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "kkcluster.infrastructure.cluster.x-k8s.io"
)

// KKClusterSpec defines the desired state of KKCluster.
type KKClusterSpec struct {
	// Distribution represents the Kubernetes distribution type of the cluster.
	Distribution string `json:"distribution,omitempty"`

	// NodeSelectorMode is the select mode of the node selector.
	// +optional
	NodeSelectorMode NodeSelectorMode `json:"nodeSelectorMode,omitempty"`

	// PipelineTemplate defines how to create a pipeline, which is used to execute all CAPKK jobs.
	PipelineTemplate PipelineTemplate `json:"pipelineTemplate,omitempty"`

	// InventoryRef is the reference of Inventory.
	InventoryRef *corev1.ObjectReference `json:"inventoryRef,omitempty"`

	// ControlPlaneLoadBalancer is optional configuration for customizing control plane behavior.
	// +optional
	ControlPlaneLoadBalancer *KKLoadBalancerSpec `json:"controlPlaneLoadBalancer,omitempty"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint"`
}

// KKClusterStatus defines the observed state of KKCluster.
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
	FailureReason *errors.ClusterStatusError `json:"failureReason,omitempty"`

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

	// Conditions defines current service state of the KKCluster.
	// +optional
	Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// PipelineTemplate defines how to create a pipeline, which is used to execute all CAPKK jobs.
type PipelineTemplate struct {
	// Project is storage for executable packages
	// +optional
	Project kkcorev1.PipelineProject `json:"project,omitempty"`
	// InventoryRef is the node configuration for playbook
	// +optional
	InventoryRef *corev1.ObjectReference `json:"inventoryRef,omitempty"`
	// ConfigRef is the global variable configuration for playbook
	// +optional
	ConfigRef *corev1.ObjectReference `json:"configRef,omitempty"`
	// Tags is the tags of playbook which to execute
	// +optional
	Tags []string `json:"tags,omitempty"`
	// SkipTags is the tags of playbook which skip execute
	// +optional
	SkipTags []string `json:"skipTags,omitempty"`
	// If Debug mode is true, It will retain runtime data after a successful execution of Pipeline,
	// which includes task execution status and parameters.
	// +optional
	Debug bool `json:"debug,omitempty"`
	// when execute in kubernetes, pipeline will create ob or cornJob to execute.
	// +optional
	JobSpec kkcorev1.PipelineJobSpec `json:"jobSpec,omitempty"`
}

// KKLoadBalancerSpec defines the desired state of an KK load balancer.
type KKLoadBalancerSpec struct {
	// The hostname on which the API server is serving.
	Host string `json:"host,omitempty"`
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
