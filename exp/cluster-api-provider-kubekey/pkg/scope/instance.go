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
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg"
)

// InstanceScopeParams defines the input parameters used to create a new InstanceScope.
type InstanceScopeParams struct {
	Client       client.Client
	Cluster      *clusterv1.Cluster
	InfraCluster *ClusterScope
	//Machine      *clusterv1.Machine
	KKMachine  *infrav1.KKMachine
	KKInstance *infrav1.KKInstance
}

// NewInstanceScope creates a new InstanceScope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewInstanceScope(params InstanceScopeParams) (*InstanceScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a InstanceScope")
	}
	//if params.Machine == nil {
	//	return nil, errors.New("machine is required when creating a InstanceScope")
	//}
	if params.Cluster == nil {
		return nil, errors.New("cluster is required when creating a InstanceScope")
	}
	if params.InfraCluster == nil {
		return nil, errors.New("kk cluster is required when creating a InstanceScope")
	}
	if params.KKMachine == nil {
		return nil, errors.New("kk machine is required when creating a InstanceScope")
	}
	if params.KKInstance == nil {
		return nil, errors.New("kk instance is required when creating a InstanceScope")
	}

	helper, err := patch.NewHelper(params.KKInstance, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	return &InstanceScope{
		client:      params.Client,
		patchHelper: helper,

		Cluster: params.Cluster,
		//Machine:      params.Machine,
		InfraCluster: params.InfraCluster,
		KKMachine:    params.KKMachine,
		KKInstance:   params.KKInstance,
	}, nil
}

// InstanceScope defines a scope defined around a machine instance and its cluster.
type InstanceScope struct {
	client      client.Client
	patchHelper *patch.Helper

	Cluster      *clusterv1.Cluster
	InfraCluster pkg.ClusterScoper
	//Machine      *clusterv1.Machine
	KKMachine  *infrav1.KKMachine
	KKInstance *infrav1.KKInstance
}

func (i *InstanceScope) InternalAddress() string {
	return i.KKInstance.Spec.InternalAddress
}

// PatchObject persists the machine spec and status.
func (i *InstanceScope) PatchObject() error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding during the deletion process).
	applicableConditions := []clusterv1.ConditionType{
		infrav1.InstanceReadyCondition,
	}

	conditions.SetSummary(i.KKInstance,
		conditions.WithConditions(applicableConditions...),
		conditions.WithStepCounterIf(i.KKInstance.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounter(),
	)

	return i.patchHelper.Patch(
		context.TODO(),
		i.KKInstance,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.InstanceReadyCondition,
		}})
}

// Close the MachineScope by updating the instance spec, instance status.
func (i *InstanceScope) Close() error {
	return i.PatchObject()
}

// HasFailed returns the failure state of the instance scope.
func (i *InstanceScope) HasFailed() bool {
	return i.KKInstance.Status.FailureReason != nil || i.KKInstance.Status.FailureMessage != nil
}

func (i *InstanceScope) SetState(state infrav1.InstanceState) {
	i.KKInstance.Status.State = state
}
