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
	"errors"
	"fmt"
	"net"

	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	infrav1beta1 "github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1beta1"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
)

const (
	// KKClusterLabelName is the label set on KKMachines and KKInstances linked to a kkCluster.
	KKClusterLabelName = "kkcluster.infrastructure.cluster.x-k8s.io/cluster-name"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	Client         ctrlclient.Client
	Cluster        *clusterv1.Cluster
	KKCluster      *infrav1beta1.KKCluster
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
		return nil, fmt.Errorf("%w: failed to init patch helper", err)
	}

	clusterScope.patchHelper = helper

	return clusterScope, nil
}

// ClusterScope defines the basic context for an actuator to operate upon.
type ClusterScope struct {
	client      ctrlclient.Client
	patchHelper *patch.Helper

	Cluster   *clusterv1.Cluster
	KKCluster *infrav1beta1.KKCluster

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
func (s *ClusterScope) GetKKMachines(ctx context.Context) (*infrav1beta1.KKMachineList, error) {
	kkMachineList := &infrav1beta1.KKMachineList{}
	if err := s.client.List(
		ctx,
		kkMachineList,
		ctrlclient.InNamespace(s.KKCluster.Namespace),
		ctrlclient.MatchingLabels{
			KKClusterLabelName: s.KKCluster.Name,
		},
	); err != nil {
		return nil, fmt.Errorf("%w: failed to list KKMachines", err)
	}

	return kkMachineList, nil
}

// ControlPlaneEndpoint returns the control plane endpoint.
func (s *ClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	return s.KKCluster.Spec.ControlPlaneEndpoint
}

// PatchClusterPhase switches `KKCluster` into the specified phase.
func (s *ClusterScope) PatchClusterPhase(ctx context.Context, phase infrav1beta1.KKClusterPhase) error {
	return s.PatchClusterWithFunc(ctx, func(kkc *infrav1beta1.KKCluster) {
		kkc.Status.Phase = phase
	})
}

// PatchClusterWithFunc executes a function that modifies the KKCluster and then patches the changes.
func (s *ClusterScope) PatchClusterWithFunc(ctx context.Context, function func(kkc *infrav1beta1.KKCluster)) error {
	// Create a deep copy of the original KKCluster to preserve its state before modification.
	expected := s.KKCluster.DeepCopy()

	// Execute the provided function to modify the KKCluster.
	function(s.KKCluster)

	// Apply the patch using the client, comparing the modified state with the original.
	if err := s.client.Status().Patch(ctx, s.KKCluster, ctrlclient.MergeFrom(expected)); err != nil {
		klog.V(5).ErrorS(err, "Update KKCluster error", "KKCluster",
			ctrlclient.ObjectKeyFromObject(s.KKCluster))

		return err
	}

	return nil
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterScope) PatchObject(ctx context.Context) error {
	return s.patchHelper.Patch(
		ctx,
		s.KKCluster,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1beta1.HostsReadyCondition,
			infrav1beta1.PreparationReadyCondition,
			infrav1beta1.EtcdReadyCondition,
			infrav1beta1.BinaryInstallCondition,
			infrav1beta1.BootstrapReadyCondition,
			infrav1beta1.ClusterReadyCondition,
		}})
}

// Close closes the current scope persisting the cluster configuration and status.
func (s *ClusterScope) Close(ctx context.Context) error {
	return s.PatchObject(ctx)
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
func (s *ClusterScope) ControlPlaneLoadBalancer() *infrav1beta1.KKLoadBalancerSpec {
	lb := s.KKCluster.Spec.ControlPlaneLoadBalancer
	if lb == nil {
		return nil
	}

	ip := net.ParseIP(lb.Host)
	if ip == nil {
		return nil
	}

	return lb
}

// APIServerPort returns the APIServerPort to use when creating the load balancer.
func (s *ClusterScope) APIServerPort() int32 {
	if s.Cluster.Spec.ClusterNetwork != nil && s.Cluster.Spec.ClusterNetwork.APIServerPort != nil {
		return *s.Cluster.Spec.ClusterNetwork.APIServerPort
	}

	return 6443
}
