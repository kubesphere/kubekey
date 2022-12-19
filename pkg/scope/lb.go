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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg"
)

// LBScope is a scope for LB.
type LBScope interface {
	pkg.ClusterScoper
	// ControlPlaneEndpoint returns KKCluster control plane endpoint
	ControlPlaneEndpoint() clusterv1.APIEndpoint
	// ControlPlaneLoadBalancer returns the KKLoadBalancerSpec
	ControlPlaneLoadBalancer() *infrav1.KKLoadBalancerSpec
	// AllInstancesInfo returns the instance info.
	AllInstancesInfo() []infrav1.InstanceInfo
}
