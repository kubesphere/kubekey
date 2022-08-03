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

package pkg

// ScopeUsage is used to indicate which controller is using a scope.
type ScopeUsage interface {
	// ControllerName returns the name of the controller that created the scope
	ControllerName() string
}

// ClusterScoper is the interface for a cluster scope.
type ClusterScoper interface {
	ScopeUsage

	// Name returns the CAPI cluster name.
	Name() string
	// Namespace returns the cluster namespace.
	Namespace() string
	// InfraClusterName returns the KKK cluster name.
	InfraClusterName() string

	// KubernetesClusterName is the name of the Kubernetes cluster.
	KubernetesClusterName() string

	// PatchObject persists the cluster configuration and status.
	PatchObject() error
	// Close closes the current scope persisting the cluster configuration and status.
	Close() error
}
