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

package cluster

import (
	"context"
	"time"

	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/remote"
	expv1 "sigs.k8s.io/cluster-api/exp/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// K3ssControlPlaneControllerName defines the controller used when creating clients.
	K3ssControlPlaneControllerName = "k3s-controlplane-controller"
)

// ManagementCluster defines all behaviors necessary for something to function as a management cluster.
type ManagementCluster interface {
	client.Reader

	GetMachinesForCluster(ctx context.Context, cluster *clusterv1.Cluster, filters ...collections.Func) (collections.Machines, error)
	GetMachinePoolsForCluster(ctx context.Context, cluster *clusterv1.Cluster) (*expv1.MachinePoolList, error)
	GetWorkloadCluster(ctx context.Context, clusterKey client.ObjectKey) (WorkloadCluster, error)
}

// Management holds operations on the management cluster.
type Management struct {
	Client  client.Reader
	Tracker *remote.ClusterCacheTracker
}

// RemoteClusterConnectionError represents a failure to connect to a remote cluster.
type RemoteClusterConnectionError struct {
	Name string
	Err  error
}

// Error satisfies the error interface.
func (e *RemoteClusterConnectionError) Error() string { return e.Name + ": " + e.Err.Error() }

// Unwrap satisfies the unwrap error inteface.
func (e *RemoteClusterConnectionError) Unwrap() error { return e.Err }

// Get implements ctrlclient.Reader
func (m *Management) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return m.Client.Get(ctx, key, obj)
}

// List implements ctrlclient.Reader
func (m *Management) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.Client.List(ctx, list, opts...)
}

// GetMachinesForCluster returns a list of machines that can be filtered or not.
// If no filter is supplied then all machines associated with the target cluster are returned.
func (m *Management) GetMachinesForCluster(ctx context.Context, cluster *clusterv1.Cluster, filters ...collections.Func) (collections.Machines, error) {
	return collections.GetFilteredMachinesForCluster(ctx, m.Client, cluster, filters...)
}

// GetMachinePoolsForCluster returns a list of machine pools owned by the cluster.
func (m *Management) GetMachinePoolsForCluster(ctx context.Context, cluster *clusterv1.Cluster) (*expv1.MachinePoolList, error) {
	selectors := []client.ListOption{
		client.InNamespace(cluster.GetNamespace()),
		client.MatchingLabels{
			clusterv1.ClusterNameLabel: cluster.GetName(),
		},
	}
	machinePoolList := &expv1.MachinePoolList{}
	err := m.Client.List(ctx, machinePoolList, selectors...)
	return machinePoolList, err
}

// GetWorkloadCluster builds a cluster object.
// The cluster comes with an etcd client generator to connect to any etcd pod living on a managed machine.
func (m *Management) GetWorkloadCluster(ctx context.Context, clusterKey client.ObjectKey) (WorkloadCluster, error) {
	restConfig, err := remote.RESTConfig(ctx, K3ssControlPlaneControllerName, m.Client, clusterKey)
	if err != nil {
		return nil, err
	}
	restConfig.Timeout = 30 * time.Second

	if m.Tracker == nil {
		return nil, errors.New("Cannot get WorkloadCluster: No remote Cluster Cache")
	}

	c, err := m.Tracker.GetClient(ctx, clusterKey)
	if err != nil {
		return nil, err
	}

	return &Workload{
		Client: c,
	}, nil
}
