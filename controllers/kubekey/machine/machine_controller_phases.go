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

package machine

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubekeyv1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha3"
)

func (r *MachineReconciler) reconcilePhase(_ context.Context, m *kubekeyv1.Machine) {
	originalPhase := m.Status.Phase //nolint:ifshort // Cannot be inlined because m.Status.Phase might be changed before it is used in the if.

	// Set the phase to "pending" if nil.
	if m.Status.Phase == "" {
		m.Status.SetTypedPhase(kubekeyv1.MachinePhasePending)
	}

	// Set the phase to "running" if there is a NodeRef field and ssh is connected.
	if m.Status.NodeRef != nil && m.Status.SSHReady {
		m.Status.SetTypedPhase(kubekeyv1.MachinePhaseRunning)
	}

	// Set the phase to "failed" if any of Status.FailureReason or Status.FailureMessage is not-nil.
	if m.Status.FailureReason != nil || m.Status.FailureMessage != nil {
		m.Status.SetTypedPhase(kubekeyv1.MachinePhaseFailed)
	}

	// Set the phase to "deleting" if the deletion timestamp is set.
	if !m.DeletionTimestamp.IsZero() {
		m.Status.SetTypedPhase(kubekeyv1.MachinePhaseDeleting)
	}

	// If the phase has changed, update the LastUpdated timestamp
	if m.Status.Phase != originalPhase {
		now := metav1.Now()
		m.Status.LastUpdated = &now
	}
}
