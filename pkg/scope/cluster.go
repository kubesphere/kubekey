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

	"github.com/go-logr/logr"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/klogr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/rootfs"
)

// ClusterScopeParams defines the input parameters used to create a new Scope.
type ClusterScopeParams struct {
	Client         client.Client
	Logger         *logr.Logger
	Cluster        *clusterv1.Cluster
	KKCluster      *infrav1.KKCluster
	ControllerName string
	RootFsBasePath string
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

	if params.Logger == nil {
		log := klogr.New()
		params.Logger = &log
	}

	clusterScope := &ClusterScope{
		Logger:         *params.Logger,
		client:         params.Client,
		Cluster:        params.Cluster,
		KKCluster:      params.KKCluster,
		controllerName: params.ControllerName,
		rootFs:         rootfs.NewLocalRootFs(params.KKCluster.Name, params.RootFsBasePath),
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
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	Cluster   *clusterv1.Cluster
	KKCluster *infrav1.KKCluster

	controllerName string

	rootFs rootfs.Interface
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

// ControlPlaneEndpoint returns the control plane endpoint.
func (s *ClusterScope) ControlPlaneEndpoint() clusterv1.APIEndpoint {
	return s.KKCluster.Spec.ControlPlaneEndpoint
}

// GlobalRegistry returns the global registry spec.
func (s *ClusterScope) GlobalRegistry() *infrav1.Registry {
	return &s.KKCluster.Spec.Registry
}

// GlobalAuth returns the global auth spec.
func (s *ClusterScope) GlobalAuth() *infrav1.Auth {
	return &s.KKCluster.Spec.Nodes.Auth
}

// AllInstancesInfo returns the all instance specs.
func (s *ClusterScope) AllInstancesInfo() []infrav1.InstanceInfo {
	return s.KKCluster.Spec.Nodes.Instances
}

// GetInstancesSpecByRole returns the KKInstance spec for the given role.
func (s *ClusterScope) GetInstancesSpecByRole(role infrav1.Role) []infrav1.KKInstanceSpec {
	var arr []infrav1.KKInstanceSpec
	for i := range s.KKCluster.Spec.Nodes.Instances {
		instance := s.KKCluster.Spec.Nodes.Instances[i]
		for _, r := range instance.Roles {
			if r == role {
				spec := infrav1.KKInstanceSpec{}
				err := copier.Copy(&spec, &instance)
				if err != nil {
					s.Logger.Info("Failed to copy instance spec", "instance", instance, "error", err)
					continue
				}
				arr = append(arr, spec)
			}
		}
	}
	return arr
}

// AllInstances returns all existing KKInstances in the cluster.
func (s *ClusterScope) AllInstances() ([]*infrav1.KKInstance, error) {
	// Get all KKInstances linked to this KKCluster.
	allInstances := &infrav1.KKInstanceList{}
	err := s.client.List(
		context.TODO(),
		allInstances,
		client.InNamespace(s.KKCluster.Namespace),
		client.MatchingLabels(s.KKCluster.Labels),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list kkInstances")
	}

	// Filter out irrelevant instances (deleting/mismatch labels) and claim orphaned instances.
	filteredInstances := make([]*infrav1.KKInstance, 0, len(allInstances.Items))
	for idx := range allInstances.Items {
		instance := &allInstances.Items[idx]
		if shouldExcludeInstance(s.KKCluster, instance) {
			continue
		}
		filteredInstances = append(filteredInstances, instance)
	}
	return filteredInstances, nil
}

// shouldExcludeInstance returns true if the instance should be filtered out, false otherwise.
func shouldExcludeInstance(cluster *infrav1.KKCluster, instance *infrav1.KKInstance) bool {
	if metav1.GetControllerOf(instance) != nil && !capiutil.IsOwnedByObject(instance, cluster) {
		return true
	}

	return false
}

// PatchObject persists the cluster configuration and status.
func (s *ClusterScope) PatchObject() error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding during the deletion process).
	applicableConditions := []clusterv1.ConditionType{
		infrav1.HostReadyCondition,
		infrav1.ExternalLoadBalancerReadyCondition,
	}

	conditions.SetSummary(s.KKCluster,
		conditions.WithConditions(applicableConditions...),
		conditions.WithStepCounterIf(s.KKCluster.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounter(),
	)

	return s.patchHelper.Patch(
		context.TODO(),
		s.KKCluster,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.HostReadyCondition,
			infrav1.ExternalLoadBalancerReadyCondition,
			infrav1.PrincipalPreparedCondition,
		}})
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

// RootFs returns the CAPKK rootfs interface.
func (s *ClusterScope) RootFs() rootfs.Interface {
	return s.rootFs
}

// ComponentZone returns the KKCluster binaries zone.
func (s *ClusterScope) ComponentZone() string {
	if s.KKCluster.Spec.Component == nil {
		return ""
	}
	return s.KKCluster.Spec.Component.ZONE
}

// ComponentHost returns the KKCluster binaries host.
func (s *ClusterScope) ComponentHost() string {
	if s.KKCluster.Spec.Component == nil {
		return ""
	}
	return s.KKCluster.Spec.Component.Host
}

// ComponentOverrides returns the KKCluster binaries overrides.
func (s *ClusterScope) ComponentOverrides() []infrav1.Override {
	if s.KKCluster.Spec.Component == nil {
		return []infrav1.Override{}
	}
	return s.KKCluster.Spec.Component.Overrides
}

// ControlPlaneLoadBalancer returns the KKLoadBalancerSpec.
func (s *ClusterScope) ControlPlaneLoadBalancer() *infrav1.KKLoadBalancerSpec {
	return s.KKCluster.Spec.ControlPlaneLoadBalancer
}

// APIServerPort returns the APIServerPort to use when creating the load balancer.
func (s *ClusterScope) APIServerPort() int32 {
	if s.Cluster.Spec.ClusterNetwork != nil && s.Cluster.Spec.ClusterNetwork.APIServerPort != nil {
		return *s.Cluster.Spec.ClusterNetwork.APIServerPort
	}
	return 6443
}
