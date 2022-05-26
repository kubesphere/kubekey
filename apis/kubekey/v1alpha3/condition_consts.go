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

package v1alpha3

// Common ConditionTypes used by Cluster API objects.
const (
	// ReadyCondition defines the Ready condition type that summarizes the operational state of a Cluster API object.
	ReadyCondition ConditionType = "Ready"
)

const (
	// MachineOwnerRemediatedCondition is set on machines that have failed a healthcheck by the MachineHealthCheck controller.
	// MachineOwnerRemediatedCondition is set to False after a health check fails, but should be changed to True by the owning controller after remediation succeeds.
	MachineOwnerRemediatedCondition ConditionType = "OwnerRemediated"

	// WaitingForRemediationReason is the reason used when a machine fails a health check and remediation is needed.
	WaitingForRemediationReason = "WaitingForRemediation"

	// RemediationFailedReason is the reason used when a remediation owner fails to remediate an unhealthy machine.
	RemediationFailedReason = "RemediationFailed"

	// RemediationInProgressReason is the reason used when an unhealthy machine is being remediated by the remediation owner.
	RemediationInProgressReason = "RemediationInProgress"

	// ExternalRemediationTemplateAvailable is set on machinehealthchecks when MachineHealthCheck controller uses external remediation.
	// ExternalRemediationTemplateAvailable is set to false if external remediation template is not found.
	ExternalRemediationTemplateAvailable ConditionType = "ExternalRemediationTemplateAvailable"

	// ExternalRemediationTemplateNotFound is the reason used when a machine health check fails to find external remediation template.
	ExternalRemediationTemplateNotFound = "ExternalRemediationTemplateNotFound"

	// ExternalRemediationRequestAvailable is set on machinehealthchecks when MachineHealthCheck controller uses external remediation.
	// ExternalRemediationRequestAvailable is set to false if creating external remediation request fails.
	ExternalRemediationRequestAvailable ConditionType = "ExternalRemediationRequestAvailable"

	// ExternalRemediationRequestCreationFailed is the reason used when a machine health check fails to create external remediation request.
	ExternalRemediationRequestCreationFailed = "ExternalRemediationRequestCreationFailed"
)

const (
	// MachinesCreatedCondition documents that the machines controlled by the MachineSet are created.
	// When this condition is false, it indicates that there was an error when cloning the infrastructure/bootstrap template or
	// when generating the machine object.
	MachinesCreatedCondition ConditionType = "MachinesCreated"

	// MachinesReadyCondition reports an aggregate of current status of the machines controlled by the MachineSet.
	MachinesReadyCondition ConditionType = "MachinesReady"

	// MachineCreationFailedReason (Severity=Error) documents a MachineSet failing to
	// generate a machine object.
	MachineCreationFailedReason = "MachineCreationFailed"

	// ResizedCondition documents a MachineSet is resizing the set of controlled machines.
	ResizedCondition ConditionType = "Resized"

	// ScalingUpReason (Severity=Info) documents a MachineSet is increasing the number of replicas.
	ScalingUpReason = "ScalingUp"

	// ScalingDownReason (Severity=Info) documents a MachineSet is decreasing the number of replicas.
	ScalingDownReason = "ScalingDown"
)
