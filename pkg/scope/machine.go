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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/klog/v2/klogr"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
)

// MachineScopeParams defines the input parameters used to create a new MachineScope.
type MachineScopeParams struct {
	Client       client.Client
	Logger       *logr.Logger
	Cluster      *clusterv1.Cluster
	InfraCluster *ClusterScope
	Machine      *clusterv1.Machine
	KKMachine    *infrav1.KKMachine
}

// NewMachineScope creates a new MachineScope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachineScope(params MachineScopeParams) (*MachineScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachineScope")
	}
	if params.Machine == nil {
		return nil, errors.New("machine is required when creating a MachineScope")
	}
	if params.Cluster == nil {
		return nil, errors.New("cluster is required when creating a MachineScope")
	}
	if params.InfraCluster == nil {
		return nil, errors.New("kk cluster is required when creating a MachineScope")
	}
	if params.KKMachine == nil {
		return nil, errors.New("kk machine is required when creating a MachineScope")
	}

	if params.Logger == nil {
		log := klogr.New()
		params.Logger = &log
	}

	helper, err := patch.NewHelper(params.KKMachine, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}
	return &MachineScope{
		Logger:      *params.Logger,
		client:      params.Client,
		patchHelper: helper,

		Cluster:      params.Cluster,
		Machine:      params.Machine,
		InfraCluster: params.InfraCluster,
		KKMachine:    params.KKMachine,
	}, nil
}

// MachineScope defines a scope defined around a machine and its cluster.
type MachineScope struct {
	logr.Logger
	client      client.Client
	patchHelper *patch.Helper

	Cluster      *clusterv1.Cluster
	InfraCluster *ClusterScope
	Machine      *clusterv1.Machine
	KKMachine    *infrav1.KKMachine
}

// Name returns the KKMachine name.
func (m *MachineScope) Name() string {
	return m.KKMachine.Name
}

// Namespace returns the namespace name.
func (m *MachineScope) Namespace() string {
	return m.KKMachine.Namespace
}

// IsControlPlane returns true if the machine is a control plane.
func (m *MachineScope) IsControlPlane() bool {
	return util.IsControlPlaneMachine(m.Machine)
}

// GetProviderID returns the KKMachine providerID from the spec.
func (m *MachineScope) GetProviderID() string {
	if m.KKMachine.Spec.ProviderID != nil {
		return *m.KKMachine.Spec.ProviderID
	}
	return ""
}

// GetInstanceID returns the KKMachine instance id by parsing Spec.ProviderID.
func (m *MachineScope) GetInstanceID() *string {
	parsed, err := noderefutil.NewProviderID(m.GetProviderID())
	if err != nil {
		return nil
	}
	return pointer.StringPtr(parsed.ID())
}

// SetProviderID sets the KKMachine providerID in spec.
func (m *MachineScope) SetProviderID(instanceID, clusterName string) {
	providerID := fmt.Sprintf("kk:///%s/%s", clusterName, instanceID)
	m.KKMachine.Spec.ProviderID = pointer.StringPtr(providerID)
}

// SetInstanceID sets the KKMachine instanceID in spec.
func (m *MachineScope) SetInstanceID(instanceID string) {
	m.KKMachine.Spec.InstanceID = pointer.StringPtr(instanceID)
}

// SetAddresses sets the KKMachine address status.
func (m *MachineScope) SetAddresses(addrs []clusterv1.MachineAddress) {
	m.KKMachine.Status.Addresses = addrs
}

// GetRoles returns the KKMachine roles.
func (m *MachineScope) GetRoles() []infrav1.Role {
	return m.KKMachine.Spec.Roles
}

// IsRole returns true if the machine has the given role.
func (m *MachineScope) IsRole(role infrav1.Role) bool {
	for _, r := range m.KKMachine.Spec.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetInstanceState returns the KKMachine instance state from the status.
func (m *MachineScope) GetInstanceState() *infrav1.InstanceState {
	return m.KKMachine.Status.InstanceState
}

// SetInstanceState sets the KKMachine status instance state.
func (m *MachineScope) SetInstanceState(v infrav1.InstanceState) {
	m.KKMachine.Status.InstanceState = &v
}

// SetReady sets the KKMachine Ready Status.
func (m *MachineScope) SetReady() {
	m.KKMachine.Status.Ready = true
}

// SetNotReady sets the KKMachine Ready Status to false.
func (m *MachineScope) SetNotReady() {
	m.KKMachine.Status.Ready = false
}

// SetFailureMessage sets the KKMachine status failure message.
func (m *MachineScope) SetFailureMessage(v error) {
	m.KKMachine.Status.FailureMessage = pointer.StringPtr(v.Error())
}

// SetFailureReason sets the KKMachine status failure reason.
func (m *MachineScope) SetFailureReason(v capierrors.MachineStatusError) {
	m.KKMachine.Status.FailureReason = &v
}

// PatchObject persists the machine spec and status.
func (m *MachineScope) PatchObject() error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding during the deletion process).
	applicableConditions := []clusterv1.ConditionType{
		infrav1.InstanceReadyCondition,
	}

	conditions.SetSummary(m.KKMachine,
		conditions.WithConditions(applicableConditions...),
		conditions.WithStepCounterIf(m.KKMachine.ObjectMeta.DeletionTimestamp.IsZero()),
		conditions.WithStepCounter(),
	)

	return m.patchHelper.Patch(
		context.TODO(),
		m.KKMachine,
		patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
			clusterv1.ReadyCondition,
			infrav1.InstanceReadyCondition,
		}})
}

// Close the MachineScope by updating the machine spec, machine status.
func (m *MachineScope) Close() error {
	return m.PatchObject()
}

// HasFailed returns the failure state of the machine scope.
func (m *MachineScope) HasFailed() bool {
	return m.KKMachine.Status.FailureReason != nil || m.KKMachine.Status.FailureMessage != nil
}

// InstanceIsInKnownState checks if the machine scope's instance state is known.
func (m *MachineScope) InstanceIsInKnownState() bool {
	state := m.GetInstanceState()
	return state != nil && infrav1.InstanceKnownStates.Has(string(*state))
}
