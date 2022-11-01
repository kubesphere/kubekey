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

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
)

// K3sControlPlaneTemplateSpec defines the desired state of K3sControlPlaneTemplate
type K3sControlPlaneTemplateSpec struct {
	Template K3sControlPlaneTemplateResource `json:"template"`
}

// K3sControlPlaneTemplateResource describes the data needed to create a K3sControlPlane from a template.
type K3sControlPlaneTemplateResource struct {
	Spec K3sControlPlaneTemplateResourceSpec `json:"spec"`
}

// K3sControlPlaneTemplateResourceSpec defines the desired state of K3sControlPlane.
// NOTE: K3sControlPlaneTemplateResourceSpec is similar to K3sControlPlaneSpec but
// omits Replicas and Version fields. These fields do not make sense on the K3sControlPlaneTemplate,
// because they are calculated by the Cluster topology reconciler during reconciliation and thus cannot
// be configured on the K3sControlPlaneTemplate.
type K3sControlPlaneTemplateResourceSpec struct {
	// MachineTemplate contains information about how machines
	// should be shaped when creating or updating a control plane.
	// +optional
	MachineTemplate *K3sControlPlaneTemplateMachineTemplate `json:"machineTemplate,omitempty"`

	// K3sConfigSpec is a K3sConfigSpec
	// to use for initializing and joining machines to the control plane.
	K3sConfigSpec infrabootstrapv1.K3sConfigSpec `json:"k3sConfigSpec"`

	// RolloutAfter is a field to indicate a rollout should be performed
	// after the specified time even if no changes have been made to the
	// K3sControlPlane.
	//
	// +optional
	RolloutAfter *metav1.Time `json:"rolloutAfter,omitempty"`

	// The RolloutStrategy to use to replace control plane machines with
	// new ones.
	// +optional
	// +kubebuilder:default={type: "RollingUpdate", rollingUpdate: {maxSurge: 1}}
	RolloutStrategy *RolloutStrategy `json:"rolloutStrategy,omitempty"`
}

// K3sControlPlaneTemplateMachineTemplate defines the template for Machines
// in a K3sControlPlaneTemplate object.
// NOTE: K3sControlPlaneTemplateMachineTemplate is similar to K3sControlPlaneMachineTemplate but
// omits ObjectMeta and InfrastructureRef fields. These fields do not make sense on the K3sControlPlaneTemplate,
// because they are calculated by the Cluster topology reconciler during reconciliation and thus cannot
// be configured on the K3sControlPlaneTemplate.
type K3sControlPlaneTemplateMachineTemplate struct {
	// NodeDrainTimeout is the total amount of time that the controller will spend on draining a controlplane node
	// The default value is 0, meaning that the node can be drained without any time limitations.
	// NOTE: NodeDrainTimeout is different from `kubectl drain --timeout`
	// +optional
	NodeDrainTimeout *metav1.Duration `json:"nodeDrainTimeout,omitempty"`

	// NodeDeletionTimeout defines how long the machine controller will attempt to delete the Node that the Machine
	// hosts after the Machine is marked for deletion. A duration of 0 will retry deletion indefinitely.
	// If no value is provided, the default value for this property of the Machine resource will be used.
	// +optional
	NodeDeletionTimeout *metav1.Duration `json:"nodeDeletionTimeout,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// K3sControlPlaneTemplate is the Schema for the k3scontrolplanetemplates API
type K3sControlPlaneTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec K3sControlPlaneTemplateSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// K3sControlPlaneTemplateList contains a list of K3sControlPlaneTemplate
type K3sControlPlaneTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3sControlPlaneTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3sControlPlaneTemplate{}, &K3sControlPlaneTemplateList{})
}
