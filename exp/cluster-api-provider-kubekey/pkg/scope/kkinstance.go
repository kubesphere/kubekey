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

package scope

import (
	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg"
)

type KKInstanceScope interface {
	pkg.ClusterScoper
	// GlobalAuth returns the global auth configuration of all instances.
	GlobalAuth() *infrav1.Auth
	// GlobalContainerManager returns the global container manager configuration of all instances.
	GlobalContainerManager() *infrav1.ContainerManager
	GlobalRegistry() *infrav1.Registry
	// AllInstancesSpec returns the KKInstanceSpec
	AllInstancesSpec() []infrav1.KKInstanceSpec
	// GetInstancesSpecByRole returns all instances filtered by role.
	GetInstancesSpecByRole(role infrav1.Role) []infrav1.KKInstanceSpec
	// AllInstances returns all KKInstance existing in cluster.
	AllInstances() ([]*infrav1.KKInstance, error)
}
