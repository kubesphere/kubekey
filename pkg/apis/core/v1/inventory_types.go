/*
Copyright 2023 The KubeSphere Authors.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// InventoryHost of Inventory
type InventoryHost map[string]runtime.RawExtension

// InventoryGroup of Inventory
type InventoryGroup struct {
	Groups []string             `json:"groups,omitempty"`
	Hosts  []string             `json:"hosts,omitempty"`
	Vars   runtime.RawExtension `json:"vars,omitempty"`
}

// InventorySpec of Inventory
type InventorySpec struct {
	// Hosts is all nodes
	Hosts InventoryHost `json:"hosts,omitempty"`
	// Vars for all host. the priority for vars is: host vars > group vars > inventory vars
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Vars runtime.RawExtension `json:"vars,omitempty"`
	// Groups nodes. a group contains repeated nodes
	// +optional
	Groups map[string]InventoryGroup `json:"groups,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

// Inventory store hosts vars for playbook.
type Inventory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InventorySpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InventoryList of Inventory
type InventoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Inventory `json:"items"`
}

// GetHostsFromGroup flatten a specific `Inventory` group with de-duplication.
func GetHostsFromGroup(inv *Inventory, groupName string, unavailableHosts, unavailableGroups map[string]struct{}) []string {
	var hosts = make([]string, 0)
	if v, ok := inv.Spec.Groups[groupName]; ok {
		unavailableGroups[groupName] = struct{}{}
		for _, cg := range v.Groups {
			if _, exist := unavailableGroups[cg]; !exist {
				unavailableGroups[cg] = struct{}{}
				hosts = append(hosts, GetHostsFromGroup(inv, cg, unavailableHosts, unavailableGroups)...)
			}
		}

		validHosts := make([]string, 0)
		for _, hostName := range v.Hosts {
			if _, ok := inv.Spec.Hosts[hostName]; ok {
				if _, exist := unavailableHosts[hostName]; !exist {
					unavailableHosts[hostName] = struct{}{}
					validHosts = append(validHosts, hostName)
				}
			}
		}
		hosts = append(hosts, validHosts...)
	}

	return hosts
}

func init() {
	SchemeBuilder.Register(&Inventory{}, &InventoryList{})
}
