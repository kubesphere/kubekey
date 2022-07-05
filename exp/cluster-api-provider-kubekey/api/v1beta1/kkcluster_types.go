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
	"fmt"
	"net"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// ClusterFinalizer allows ReconcileKKCluster to clean up KK resources associated with KKCluster before
	// removing it from the apiserver.
	ClusterFinalizer = "kkcluster.infrastructure.cluster.x-k8s.io"
)

// KKClusterSpec defines the desired state of KKCluster
type KKClusterSpec struct {
	//// Selector is a label query over machines that should match the replica count.
	//// Label keys and values that must match in order to be controlled by this MachineSet.
	//// It must match the machine template's labels.
	//// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	//Selector metav1.LabelSelector `json:"selector"`

	Nodes Nodes `json:"nodes"`

	// ControlPlaneEndpoint represents the endpoint used to communicate with the control plane.
	// +optional
	ControlPlaneEndpoint ControlPlaneEndPoint `json:"controlPlaneEndpoint"`

	// Registry represents the cluster image registry used to pull the images.
	// +optional
	Registry Registry `json:"registry"`
}

type ControlPlaneEndPoint struct {
	// +optional
	Address string `json:"address"`

	Domain string `json:"domain"`

	// The port on which the API server is serving.
	Port int32 `json:"port"`
}

type Nodes struct {
	// Auth is the SSH authentication information of all instance. It is a global auth configuration.
	// +optional
	Auth Auth `json:"auth"`

	// ContainerManager is the container manager config of all instance. It is a global container manager configuration.
	// +optional
	ContainerManager ContainerManager `json:"containerManager"`

	// Instances defines all instance contained in kkcluster.
	Instances []KKInstanceSpec `json:"instances"`
}

// KKClusterStatus defines the observed state of KKCluster
type KKClusterStatus struct {
	// +kubebuilder:default=false
	Ready          bool                     `json:"ready"`
	FailureDomains clusterv1.FailureDomains `json:"failureDomains,omitempty"`
	Conditions     clusterv1.Conditions     `json:"conditions,omitempty"`
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

// IsZero returns true if both host and port are zero values.
func (c ControlPlaneEndPoint) IsZero() bool {
	return c.Address == "" && c.Port == 0
}

// IsValid returns true if both host and port are non-zero values.
func (c ControlPlaneEndPoint) IsValid() bool {
	return c.Address != "" && c.Port != 0
}

// String returns a formatted version HOST:PORT of this APIEndpoint.
func (c ControlPlaneEndPoint) String() string {
	return net.JoinHostPort(c.Address, fmt.Sprintf("%d", c.Port))
}
