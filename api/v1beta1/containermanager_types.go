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

// Default values.
const (
	DockerType             = "docker"
	DefaultDockerVersion   = "20.10.8"
	DefaultDockerCRISocket = ""

	ContainerdType             = "containerd"
	DefaultContainerdVersion   = "1.6.4"
	DefaultContainerdCRISocket = "unix:///var/run/containerd/containerd.sock"
)

// ContainerManager defines the desired state of ContainerManager
type ContainerManager struct {
	// CRISocket is used to connect an existing CRIClient.
	// +optional
	CRISocket string `json:"criSocket,omitempty"`

	// Type defines the type of ContainerManager.
	// "docker", "containerd"
	Type string `json:"type,omitempty"`

	// Version defines the version of ContainerManager.
	Version string `json:"version,omitempty"`
}
