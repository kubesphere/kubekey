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
	"context"
	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1alpha1"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// KKClusterLabelName is the label set on KKMachines and KKInstances linked to a kkCluster.
	KKClusterLabelName = "kkcluster.infrastructure.cluster.x-k8s.io/cluster-name"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	Client         client.Client
	Cluster        *clusterv1.Cluster
	KKCluster      *v1alpha1.KKCluster
	ControllerName string
}

// NewClusterScope creates a new Scope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewClusterScope(params ClusterScopeParams) (*ClusterScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.KKCluster == nil {
		return nil, errors.New("failed to generate new scope from nil KKCluster")
	}

	clusterScope := &ClusterScope{
		client:         params.Client,
		Cluster:        params.Cluster,
		KKCluster:      params.KKCluster,
		controllerName: params.ControllerName,
	}

	helper, err := patch.NewHelper(params.KKCluster, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	clusterScope.patchHelper = helper

	return clusterScope, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	client      client.Client
	patchHelper *patch.Helper

	Cluster   *clusterv1.Cluster
	KKCluster *v1alpha1.KKCluster

	controllerName string
}

// Name returns the CAPI cluster name.
func (s *ClusterScope) Name() string {
	return s.Cluster.Name
}

// Namespace returns the cluster namespace.
func (s *ClusterScope) Namespace() string {
	return s.Cluster.Namespace
}

// InfraClusterName returns the KK cluster name.
func (s *ClusterScope) InfraClusterName() string {
	return s.KKCluster.Name
}

// KubernetesClusterName is the name of the Kubernetes cluster.
func (s *ClusterScope) KubernetesClusterName() string {
	return s.Cluster.Name
}

// GetKKMachines returns the list of KKMachines for a KKCluster.
func (s *ClusterScope) GetKKMachines(ctx context.Context) (*v1alpha1.KKMachineList, error) {
	kkMachineList := &v1alpha1.KKMachineList{}
	if err := s.client.List(
		ctx,
		kkMachineList,
		client.InNamespace(s.KKCluster.Namespace),
		client.MatchingLabels{
			KKClusterLabelName: s.KKCluster.Name,
		},
	); err != nil {
		return nil, errors.Wrap(err, "failed to list KKMachines")
	}

	return kkMachineList, nil
}

// GetMachines returns the collections of machines for a KKCluster.
// func (s *ClusterScope) GetMachines(ctx context.Context, filters ...capicollections.Func) (capicollections.Machines, error) {
//	kkml, err := s.GetKKMachines(ctx)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to get KKMachines")
//	}
//
//	machines := make(capicollections.Machines, len(kkml.Items))
//	for i := range kkml.Items {
//		machine, err := capiutil.GetOwnerMachine(ctx, s.client, kkml.Items[i].ObjectMeta)
//		if err != nil {
//			return nil, errors.Wrap(err, "failed to get owner machine")
//		}
//		machines.Insert(machine)
//	}
//	return machines.Filter(filters...), nil
// }

// ControlPlaneEndpoint returns the control plane endpoint.
func (s *ClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	return s.KKCluster.Spec.ControlPlaneEndpoint
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterScope) PatchObject() error {
	return s.patchHelper.Patch(
		context.TODO(),
		s.KKCluster)
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close() error {
	return s.PatchObject()
}

// ControllerName returns the name of the controller that
// created the ClusterScope.
func (s *ClusterScope) ControllerName() string {
	return s.controllerName
}

// Distribution returns Kubernetes distribution of the cluster.
func (s *ClusterScope) Distribution() string {
	return s.KKCluster.Spec.Distribution
}

// ControlPlaneLoadBalancer returns the KKLoadBalancerSpec.
func (s *ClusterScope) ControlPlaneLoadBalancer() *v1alpha1.KKLoadBalancerSpec {
	return s.KKCluster.Spec.ControlPlaneLoadBalancer
}

// APIServerPort returns the APIServerPort to use when creating the load balancer.
func (s *ClusterScope) APIServerPort() int32 {
	if s.Cluster.Spec.ClusterNetwork != nil && s.Cluster.Spec.ClusterNetwork.APIServerPort != nil {
		return *s.Cluster.Spec.ClusterNetwork.APIServerPort
	}
	return 6443
}
