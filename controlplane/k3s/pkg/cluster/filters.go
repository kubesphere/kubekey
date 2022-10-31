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
	"encoding/json"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
	infracontrolplanev1 "github.com/kubesphere/kubekey/controlplane/k3s/api/v1beta1"
)

// MatchesMachineSpec returns a filter to find all machines that matches with KCP config and do not require any rollout.
// Kubernetes version, infrastructure template, and K3sConfig field need to be equivalent.
func MatchesMachineSpec(infraConfigs map[string]*unstructured.Unstructured, machineConfigs map[string]*infrabootstrapv1.K3sConfig, kcp *infracontrolplanev1.K3sControlPlane) func(machine *clusterv1.Machine) bool {
	return collections.And(
		func(machine *clusterv1.Machine) bool {
			return matchMachineTemplateMetadata(kcp, machine)
		},
		collections.MatchesKubernetesVersion(kcp.Spec.Version),
		//MatchesK3sBootstrapConfig(machineConfigs, kcp),
		MatchesTemplateClonedFrom(infraConfigs, kcp),
	)
}

// MatchesTemplateClonedFrom returns a filter to find all machines that match a given KCP infra template.
func MatchesTemplateClonedFrom(infraConfigs map[string]*unstructured.Unstructured, kcp *infracontrolplanev1.K3sControlPlane) collections.Func {
	return func(machine *clusterv1.Machine) bool {
		if machine == nil {
			return false
		}
		infraObj, found := infraConfigs[machine.Name]
		if !found {
			// Return true here because failing to get infrastructure machine should not be considered as unmatching.
			return true
		}

		clonedFromName, ok1 := infraObj.GetAnnotations()[clusterv1.TemplateClonedFromNameAnnotation]
		clonedFromGroupKind, ok2 := infraObj.GetAnnotations()[clusterv1.TemplateClonedFromGroupKindAnnotation]
		if !ok1 || !ok2 {
			// All kcp cloned infra machines should have this annotation.
			// Missing the annotation may be due to older version machines or adopted machines.
			// Should not be considered as mismatch.
			return true
		}

		// Check if the machine's infrastructure reference has been created from the current KCP infrastructure template.
		if clonedFromName != kcp.Spec.MachineTemplate.InfrastructureRef.Name ||
			clonedFromGroupKind != kcp.Spec.MachineTemplate.InfrastructureRef.GroupVersionKind().GroupKind().String() {
			return false
		}

		// Check if the machine template metadata matches with the infrastructure object.
		if !matchMachineTemplateMetadata(kcp, infraObj) {
			return false
		}
		return true
	}
}

// MatchesK3sBootstrapConfig checks if machine's K3sConfigSpec is equivalent with KCP's K3sConfigSpec.
func MatchesK3sBootstrapConfig(machineConfigs map[string]*infrabootstrapv1.K3sConfig, kcp *infracontrolplanev1.K3sControlPlane) collections.Func {
	return func(machine *clusterv1.Machine) bool {
		if machine == nil {
			return false
		}

		// Check if KCP and machine ClusterConfiguration matches, if not return
		if match := matchClusterConfiguration(kcp, machine); !match {
			return false
		}

		bootstrapRef := machine.Spec.Bootstrap.ConfigRef
		if bootstrapRef == nil {
			// Missing bootstrap reference should not be considered as unmatching.
			// This is a safety precaution to avoid selecting machines that are broken, which in the future should be remediated separately.
			return true
		}

		machineConfig, found := machineConfigs[machine.Name]
		if !found {
			// Return true here because failing to get K3sConfig should not be considered as unmatching.
			// This is a safety precaution to avoid rolling out machines if the client or the api-server is misbehaving.
			return true
		}

		// Check if the machine template metadata matches with the infrastructure object.
		if !matchMachineTemplateMetadata(kcp, machineConfig) {
			return false
		}

		// Check if KCP and machine InitConfiguration or JoinConfiguration matches
		// NOTE: only one between init configuration and join configuration is set on a machine, depending
		// on the fact that the machine was the initial control plane node or a joining control plane node.
		return matchInitOrJoinConfiguration(machineConfig, kcp)
	}
}

// matchClusterConfiguration verifies if KCP and machine ClusterConfiguration matches.
// NOTE: Machines that have K3sClusterConfigurationAnnotation will have to match with KCP ClusterConfiguration.
// If the annotation is not present (machine is either old or adopted), we won't roll out on any possible changes
// made in KCP's ClusterConfiguration given that we don't have enough information to make a decision.
// Users should use KCP.Spec.RolloutAfter field to force a rollout in this case.
func matchClusterConfiguration(kcp *infracontrolplanev1.K3sControlPlane, machine *clusterv1.Machine) bool {
	machineClusterConfigStr, ok := machine.GetAnnotations()[infracontrolplanev1.K3sServerConfigurationAnnotation]
	if !ok {
		// We don't have enough information to make a decision; don't' trigger a roll out.
		return true
	}

	machineClusterConfig := &infrabootstrapv1.Cluster{}
	// ClusterConfiguration annotation is not correct, only solution is to rollout.
	// The call to json.Unmarshal has to take a pointer to the pointer struct defined above,
	// otherwise we won't be able to handle a nil ClusterConfiguration (that is serialized into "null").
	// See https://github.com/kubernetes-sigs/cluster-api/issues/3353.
	if err := json.Unmarshal([]byte(machineClusterConfigStr), &machineClusterConfig); err != nil {
		return false
	}

	// If any of the compared values are nil, treat them the same as an empty ClusterConfiguration.
	if machineClusterConfig == nil {
		machineClusterConfig = &infrabootstrapv1.Cluster{}
	}
	kcpLocalClusterConfiguration := kcp.Spec.K3sConfigSpec.Cluster
	if kcpLocalClusterConfiguration == nil {
		kcpLocalClusterConfiguration = &infrabootstrapv1.Cluster{}
	}

	// Compare and return.
	return reflect.DeepEqual(machineClusterConfig, kcpLocalClusterConfiguration)
}

// matchInitOrJoinConfiguration verifies if KCP and machine ServerConfiguration or AgentConfiguration matches.
// NOTE: By extension this method takes care of detecting changes in other fields of the K3sConfig configuration (e.g. Files, Mounts etc.)
func matchInitOrJoinConfiguration(machineConfig *infrabootstrapv1.K3sConfig, kcp *infracontrolplanev1.K3sControlPlane) bool {
	if machineConfig == nil {
		// Return true here because failing to get K3sConfig should not be considered as unmatching.
		// This is a safety precaution to avoid rolling out machines if the client or the api-server is misbehaving.
		return true
	}

	// takes the K3sConfigSpec from KCP and applies the transformations required
	// to allow a comparison with the K3sConfig referenced from the machine.
	kcpConfig := getAdjustedKcpConfig(kcp, machineConfig)

	// Default both K3sConfigSpecs before comparison.
	// *Note* This assumes that newly added default values never
	// introduce a semantic difference to the unset value.
	// But that is something that is ensured by our API guarantees.
	infrabootstrapv1.DefaultK3sConfigSpec(kcpConfig)
	infrabootstrapv1.DefaultK3sConfigSpec(&machineConfig.Spec)

	// cleanups all the fields that are not relevant for the comparison.
	cleanupConfigFields(kcpConfig, machineConfig)

	return reflect.DeepEqual(&machineConfig.Spec, kcpConfig)
}

// getAdjustedKcpConfig takes the K3sConfigSpec from KCP and applies the transformations required
// to allow a comparison with the K3sConfig referenced from the machine.
// NOTE: The KCP controller applies a set of transformations when creating a K3sConfig referenced from the machine,
// mostly depending on the fact that the machine was the initial control plane node or a joining control plane node.
// In this function we don't have such information, so we are making the K3sConfigSpec similar to the KubeadmConfig.
func getAdjustedKcpConfig(kcp *infracontrolplanev1.K3sControlPlane, machineConfig *infrabootstrapv1.K3sConfig) *infrabootstrapv1.K3sConfigSpec {
	kcpConfig := kcp.Spec.K3sConfigSpec.DeepCopy()

	// Machine's join configuration is nil when it is the first machine in the control plane.
	if machineConfig.Spec.AgentConfiguration == nil {
		kcpConfig.AgentConfiguration = nil
	}

	// Machine's init configuration is nil when the control plane is already initialized.
	if machineConfig.Spec.ServerConfiguration == nil {
		kcpConfig.ServerConfiguration = nil
	}

	return kcpConfig
}

// cleanupConfigFields cleanups all the fields that are not relevant for the comparison.
func cleanupConfigFields(kcpConfig *infrabootstrapv1.K3sConfigSpec, machineConfig *infrabootstrapv1.K3sConfig) {
	// KCP ClusterConfiguration will only be compared with a machine's ClusterConfiguration annotation, so
	// we are cleaning up from the reflect.DeepEqual comparison.
	kcpConfig.Cluster = nil
	machineConfig.Spec.Cluster = nil

	kcpConfig.ServerConfiguration = nil
	machineConfig.Spec.ServerConfiguration = nil

	// If KCP JoinConfiguration is not present, set machine JoinConfiguration to nil (nothing can trigger rollout here).
	// NOTE: this is required because CABPK applies an empty joinConfiguration in case no one is provided.
	if kcpConfig.AgentConfiguration == nil {
		machineConfig.Spec.AgentConfiguration = nil
	}
}

// matchMachineTemplateMetadata matches the machine template object meta information,
// specifically annotations and labels, against an object.
func matchMachineTemplateMetadata(kcp *infracontrolplanev1.K3sControlPlane, obj client.Object) bool {
	// Check if annotations and labels match.
	if !isSubsetMapOf(kcp.Spec.MachineTemplate.ObjectMeta.Annotations, obj.GetAnnotations()) {
		return false
	}
	if !isSubsetMapOf(kcp.Spec.MachineTemplate.ObjectMeta.Labels, obj.GetLabels()) {
		return false
	}
	return true
}

func isSubsetMapOf(base map[string]string, existing map[string]string) bool {
loopBase:
	for key, value := range base {
		for existingKey, existingValue := range existing {
			if existingKey == key && existingValue == value {
				continue loopBase
			}
		}
		// Return false right away if a key value pair wasn't found.
		return false
	}
	return true
}
