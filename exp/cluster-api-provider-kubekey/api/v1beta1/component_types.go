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

// Component is optional configuration for modifying the FTP server.
type Component struct {
	// ZONE is the zone of the KKCluster where can get the binaries.
	// If you have problem to access https://storage.googleapis.com, you can set "zone: cn".
	// +optional
	ZONE string `json:"zone,omitempty"`

	// Host is the host to download the binaries.
	// +optional
	Host string `json:"host,omitempty"`

	// Overrides is a list of components download information that need to be overridden.
	// +optional
	Overrides []Override `json:"overrides,omitempty"`
}

// Override is a component download information that need to be overridden.
type Override struct {
	// ID is the component id name. e.g. kubeadm, kubelet, containerd, etc.
	ID string `json:"id,omitempty"`

	// Arch is the component arch. e.g. amd64, arm64, etc.
	Arch string `json:"arch,omitempty"`

	// Version is the component version. e.g. v1.21.1, v1.22.0, etc.
	Version string `json:"version,omitempty"`

	// URL is the download url of the binaries.
	URL string `json:"url,omitempty"`

	// Path defines the URL path, which is the string of information that comes after the top level domain name.
	Path string `json:"path,omitempty"`

	// Checksum is the SHA256 checksum of the binary.
	// +optional
	Checksum string `json:"checksum,omitempty"`
}
