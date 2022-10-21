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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/cluster-api/util"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelNodeRoleOldControlPlane = "node-role.kubernetes.io/master" // Deprecated: https://github.com/kubernetes/kubeadm/issues/2200
	labelNodeRoleControlPlane    = "node-role.kubernetes.io/control-plane"
)

// WorkloadCluster defines all behaviors necessary to upgrade kubernetes on a workload cluster
//
// TODO: Add a detailed description to each of these method definitions.
type WorkloadCluster interface {
	// Basic health and status checks.
	ClusterStatus(ctx context.Context) (Status, error)
	UpdateAgentConditions(ctx context.Context, controlPlane *ControlPlane)
	UpdateEtcdConditions(ctx context.Context, controlPlane *ControlPlane)
}

// Workload defines operations on workload clusters.
type Workload struct {
	Client ctrlclient.Client
}

// Status holds stats information about the cluster.
type Status struct {
	// Nodes are a total count of nodes
	Nodes int32
	// ReadyNodes are the count of nodes that are reporting ready
	ReadyNodes int32
	// HasK3sConfig will be true if the kubeadm config map has been uploaded, false otherwise.
	HasK3sConfig bool
}

func (w *Workload) getControlPlaneNodes(ctx context.Context) (*corev1.NodeList, error) {
	controlPlaneNodes := &corev1.NodeList{}
	controlPlaneNodeNames := sets.NewString()

	for _, label := range []string{labelNodeRoleOldControlPlane, labelNodeRoleControlPlane} {
		nodes := &corev1.NodeList{}
		if err := w.Client.List(ctx, nodes, ctrlclient.MatchingLabels(map[string]string{
			label: "",
		})); err != nil {
			return nil, err
		}

		for i := range nodes.Items {
			node := nodes.Items[i]

			// Continue if we already added that node.
			if controlPlaneNodeNames.Has(node.Name) {
				continue
			}

			controlPlaneNodeNames.Insert(node.Name)
			controlPlaneNodes.Items = append(controlPlaneNodes.Items, node)
		}
	}

	return controlPlaneNodes, nil
}

// ClusterStatus returns the status of the cluster.
func (w *Workload) ClusterStatus(ctx context.Context) (Status, error) {
	status := Status{}

	// count the control plane nodes
	nodes, err := w.getControlPlaneNodes(ctx)
	if err != nil {
		return status, err
	}

	for _, node := range nodes.Items {
		nodeCopy := node
		status.Nodes++
		if util.IsNodeReady(&nodeCopy) {
			status.ReadyNodes++
		}
	}

	return status, nil
}
