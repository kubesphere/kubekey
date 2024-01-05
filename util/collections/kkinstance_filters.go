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

package collections

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
)

// Func is the functon definition for a filter.
type Func func(kkInstance *infrav1.KKInstance) bool

// And returns a filter that returns true if all of the given filters returns true.
func And(filters ...Func) Func {
	return func(kkInstance *infrav1.KKInstance) bool {
		for _, f := range filters {
			if !f(kkInstance) {
				return false
			}
		}
		return true
	}
}

// Or returns a filter that returns true if any of the given filters returns true.
func Or(filters ...Func) Func {
	return func(kkInstance *infrav1.KKInstance) bool {
		for _, f := range filters {
			if f(kkInstance) {
				return true
			}
		}
		return false
	}
}

// Not returns a filter that returns the opposite of the given filter.
func Not(mf Func) Func {
	return func(kkInstance *infrav1.KKInstance) bool {
		return !mf(kkInstance)
	}
}

// OwnedKKInstances returns a filter to find all kkInstances owned by specified owner.
// Usage: GetFilteredKKInstancesForKKCluster(ctx, client, cluster, OwnedKKInstances(controlPlane)).
func OwnedKKInstances(owner client.Object) func(kkInstance *infrav1.KKInstance) bool {
	return func(kkInstance *infrav1.KKInstance) bool {
		if kkInstance == nil {
			return false
		}
		return util.IsOwnedByObject(kkInstance, owner)
	}
}

// ControlPlaneKKInstances returns a filter to find all control plane KKInstance for a cluster, regardless of ownership.
// Usage: GetFilteredKKInstancesForKKCluster(ctx, client, cluster, ControlPlaneKKInstances(cluster.Name)).
func ControlPlaneKKInstances(clusterName string) func(kkInstance *infrav1.KKInstance) bool {
	selector := ControlPlaneSelectorForCluster(clusterName)
	return func(kkInstance *infrav1.KKInstance) bool {
		if kkInstance == nil {
			return false
		}
		return selector.Matches(labels.Set(kkInstance.Labels))
	}
}

// ActiveKKInstances returns a filter to find all active kkinstances.
// Usage: GetFilteredKKInstancesForKKCluster(ctx, client, cluster, ActiveKKInstances).
func ActiveKKInstances(kkInstance *infrav1.KKInstance) bool {
	if kkInstance == nil {
		return false
	}
	return kkInstance.DeletionTimestamp.IsZero()
}

// ControlPlaneSelectorForCluster returns the label selector necessary to get control plane KKInstance for a given cluster.
func ControlPlaneSelectorForCluster(clusterName string) labels.Selector {
	must := func(r *labels.Requirement, err error) labels.Requirement {
		if err != nil {
			panic(err)
		}
		return *r
	}
	return labels.NewSelector().Add(
		must(labels.NewRequirement(clusterv1.ClusterNameLabel, selection.Equals, []string{clusterName})),
		must(labels.NewRequirement(clusterv1.MachineControlPlaneLabel, selection.Exists, []string{})),
	)
}
