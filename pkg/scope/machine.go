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

	infrastructurev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/capkk/v1beta1"

	"k8s.io/utils/ptr"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineScopeParams defines the input parameters used to create a new MachineScope.
type MachineScopeParams struct {
	Client       ctrlclient.Client
	ClusterScope *ClusterScope
	Machine      *clusterv1.Machine
	KKMachine    *infrastructurev1alpha1.KKMachine
}

// NewMachineScope creates a new MachineScope from the supplied parameters.
// This is meant to be called for each reconcile iteration.
func NewMachineScope(params MachineScopeParams) (*MachineScope, error) {
	if params.Client == nil {
		return nil, errors.New("client is required when creating a MachineScope")
	}
	if params.ClusterScope == nil {
		return nil, errors.New("cluster is required when creating a MachineScope")
	}
	if params.Machine == nil {
		return nil, errors.New("machine is required when creating a MachineScope")
	}
	if params.KKMachine == nil {
		return nil, errors.New("kk machine is required when creating a MachineScope")
	}

	helper, err := patch.NewHelper(params.KKMachine, params.Client)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to init patch helper", err)
	}

	return &MachineScope{
		client:      params.Client,
		patchHelper: helper,

		Machine:      params.Machine,
		ClusterScope: params.ClusterScope,
		KKMachine:    params.KKMachine,
	}, nil
}

// MachineScope defines a scope defined around a machine and its cluster.
type MachineScope struct {
	client      ctrlclient.Client
	patchHelper *patch.Helper

	ClusterScope *ClusterScope
	Machine      *clusterv1.Machine
	KKMachine    *infrastructurev1alpha1.KKMachine
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

// SetProviderID sets the KKMachine providerID in spec.
func (m *MachineScope) SetProviderID(kkMachineID string) {
	m.KKMachine.Spec.ProviderID = ptr.To(kkMachineID)
}

// GetRoles returns the KKMachine roles.
func (m *MachineScope) GetRoles() []string {
	return m.KKMachine.Spec.Roles
}

// IsRole returns true if the machine has the given role.
func (m *MachineScope) IsRole(role string) bool {
	for _, r := range m.KKMachine.Spec.Roles {
		if r == role {
			return true
		}
	}

	return false
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
	m.KKMachine.Status.FailureMessage = ptr.To(v.Error())
}

// SetFailureReason sets the KKMachine status failure reason.
func (m *MachineScope) SetFailureReason(v capierrors.MachineStatusError) {
	m.KKMachine.Status.FailureReason = &v
}

// PatchObject persists the machine spec and status.
func (m *MachineScope) PatchObject(ctx context.Context) error {
	// Always update the readyCondition by summarizing the state of other conditions.
	// A step counter is added to represent progress during the provisioning process (instead we are hiding during the deletion process).
	return m.patchHelper.Patch(
		ctx,
		m.KKMachine)
}

// Close the MachineScope by updating the machine spec, machine status.
func (m *MachineScope) Close(ctx context.Context) error {
	return m.PatchObject(ctx)
}

// HasFailed returns the failure state of the machine scope.
func (m *MachineScope) HasFailed() bool {
	return m.KKMachine.Status.FailureReason != nil || m.KKMachine.Status.FailureMessage != nil
}
