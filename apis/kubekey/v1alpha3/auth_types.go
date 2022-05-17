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

// Auth contains the SSH authentication configuration for machines.
type Auth struct {
	// User is the username for SSH authentication.
	User string `yaml:"user,omitempty" json:"user,omitempty"`

	// Password is the password for SSH authentication.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`

	// Port is the port for SSH authentication.
	Port int `yaml:"port,omitempty" json:"port,omitempty"`

	// PrivateKey is the value of the private key for SSH authentication.
	PrivateKey string `yaml:"privateKey,omitempty" json:"privateKey,omitempty"`

	// PrivateKeyFile is the path to the private key for SSH authentication.
	PrivateKeyPath string `yaml:"privateKeyPath,omitempty" json:"privateKeyPath,omitempty"`

	// Timeout is the timeout for establish an SSH connection.
	// +optional
	Timeout *int64 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}
