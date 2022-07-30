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

package v1beta1

import (
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// KKCluster condition
const (
	PrincipalPreparedCondition clusterv1.ConditionType = "PrincipalPrepared"
)

const (
	HostReadyCondition clusterv1.ConditionType = "HostReadyCondition"
)

const (
	// LoadBalancerReadyCondition reports on whether a control plane load balancer was successfully reconciled.
	LoadBalancerReadyCondition clusterv1.ConditionType = "LoadBalancerReady"
)

// KKMachine condition
const (
	// InstanceReadyCondition reports on current status of the SSH instance. Ready indicates the instance is in a Running state.
	InstanceReadyCondition clusterv1.ConditionType = "InstanceReady"

	// InstanceNotFoundReason used when the instance couldn't be retrieved.
	InstanceNotFoundReason = "InstanceNotFound"
	// InstanceCleanedReason instance is in a Cleared state.
	InstanceCleanedReason = "InstanceCleaned"
	// InstanceNotReadyReason used when the instance is in a pending state.
	InstanceNotReadyReason = "InstanceNotReady"
	// InstanceBootstrapStartedReason set when the provisioning of an instance started.
	InstanceBootstrapStartedReason = "InstanceBootstrapStarted"
	// InstanceBootstrapFailedReason used for failures during instance provisioning.
	InstanceBootstrapFailedReason = "InstanceBootstrapFailed"
	// WaitingForClusterInfrastructureReason used when machine is waiting for cluster infrastructure to be ready before proceeding.
	WaitingForClusterInfrastructureReason = "WaitingForClusterInfrastructure"
	// WaitingForBootstrapDataReason used when machine is waiting for bootstrap data to be ready before proceeding.
	WaitingForBootstrapDataReason = "WaitingForBootstrapData"
)

// KKInstance condition
const (
	KKInstanceBootstrappedCondition clusterv1.ConditionType = "InstanceBootstrapped"
	KKInstanceInitOSFailedReason                            = "InitOSFailed"
)

const (
	KKInstanceBinariesReadyCondition clusterv1.ConditionType = "InstanceBinariesReady"
	KKInstanceGetBinaryFailedReason                          = "GetBinaryFailed"
)

const (
	KKInstanceCRIReadyCondition      clusterv1.ConditionType = "InstanceCRIReady"
	KKInstanceInstallCRIFailedReason                         = "InstallCRIFailed"
)

const (
	KKInstanceProvisionedCondition       clusterv1.ConditionType = "InstanceProvisioned"
	KKInstanceRunCloudConfigFailedReason                         = "RunCloudConfigFailed"
)

const (
	KKInstanceDeletingBootstrapCondition   clusterv1.ConditionType = "InstanceDeletingBootstrapped"
	KKInstanceClearEnvironmentFailedReason                         = "ClearEnvironmentFailed"

	// CleaningReason (Severity=Info) documents a machine node being cleaned.
	CleaningReason = "Cleaning"
)
