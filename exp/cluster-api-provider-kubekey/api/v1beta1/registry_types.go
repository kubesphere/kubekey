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
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Registry is the configuration for a cluster registry
type Registry struct {
	metav1.TypeMeta `json:",inline"`

	// PrivateRegistry defines the private registry address of ContainerManager.
	PrivateRegistry string `json:"privateRegistry"`

	// InsecureRegistries defines the insecure registries of ContainerManager.
	InsecureRegistries []string `json:"insecureRegistries,omitempty"`

	// RegistryMirrors defines the registry mirrors of this PrivateRegistry.
	RegistryMirrors []string `json:"registryMirrors,omitempty"`

	// NamespaceOverride defines the namespace override of this PrivateRegistry.
	NamespaceOverride string `json:"namespaceOverride"`

	// Auth defines the auth of this PrivateRegistry.
	Auth RegistryAuth `json:"auth"`
}

// RegistryAuth defines the auth of a registry
type RegistryAuth struct {
	// Username defines the username of this PrivateRegistry.
	Username string `json:"username"`

	// Password defines the password of this PrivateRegistry.
	Password string `json:"password"`

	// InsecureSkipVerify allow contacting this PrivateRegistry over HTTPS with failed TLS verification.
	InsecureSkipVerify bool `json:"insecureSkipVerify"`

	// PlainHTTP allow contacting this PrivateRegistry over HTTP.
	PlainHTTP bool `json:"plainHTTP"`

	// CertsPath defines the path of the certs files of this PrivateRegistry.
	CertsPath string `json:"certsPath"`

	// CAFile is an SSL Certificate Authority file used to secure etcd communication.
	CAFile string `yaml:"caFile" json:"caFile,omitempty"`
	// CertFile is an SSL certification file used to secure etcd communication.
	CertFile string `yaml:"certFile" json:"certFile,omitempty"`
	// KeyFile is an SSL key file used to secure etcd communication.
	KeyFile string `yaml:"keyFile" json:"keyFile,omitempty"`
}
