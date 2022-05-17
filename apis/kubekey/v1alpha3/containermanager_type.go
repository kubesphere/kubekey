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

package v1alpha3

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ContainerManager defines the desired state of ContainerManager
type ContainerManager struct {
	// CRISocket is used to connect an existing CRIClient.
	// +optional
	CRISocket *string `json:"criSocket,omitempty"`

	// Type defines the type of ContainerManager.
	// "docker", "containerd"
	Type string `json:"type"`

	// Version defines the version of ContainerManager.
	Version string `json:"version"`

	// Registries defines the registries' config of ContainerManager.
	Registries []RegistryConfig `json:"registries"`
}
