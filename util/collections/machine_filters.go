/*
Copyright 2020 The Kubernetes Authors.

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
	"github.com/kubesphere/kubekey/util"
	"github.com/kubesphere/kubekey/util/conditions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha3"
)

// Func is the functon definition for a filter.
type Func func(machine *kubekeyv1.Machine) bool

// And returns a filter that returns true if all of the given filters returns true.
func And(filters ...Func) Func {
	return func(machine *kubekeyv1.Machine) bool {
		for _, f := range filters {
			if !f(machine) {
				return false
			}
		}
		return true
	}
}

// Or returns a filter that returns true if any of the given filters returns true.
func Or(filters ...Func) Func {
	return func(machine *kubekeyv1.Machine) bool {
		for _, f := range filters {
			if f(machine) {
				return true
			}
		}
		return false
	}
}

// Not returns a filter that returns the opposite of the given filter.
func Not(mf Func) Func {
	return func(machine *kubekeyv1.Machine) bool {
		return !mf(machine)
	}
}

// HasControllerRef is a filter that returns true if the machine has a controller ref.
func HasControllerRef(machine *kubekeyv1.Machine) bool {
	if machine == nil {
		return false
	}
	return metav1.GetControllerOf(machine) != nil
}

// OwnedMachines returns a filter to find all machines owned by specified owner.
// Usage: GetFilteredMachinesForCluster(ctx, client, cluster, OwnedMachines(controlPlane)).
func OwnedMachines(owner client.Object) func(machine *kubekeyv1.Machine) bool {
	return func(machine *kubekeyv1.Machine) bool {
		if machine == nil {
			return false
		}
		return util.IsOwnedByObject(machine, owner)
	}
}

// ControlPlaneMachines returns a filter to find all control plane machines for a cluster, regardless of ownership.
// Usage: GetFilteredMachinesForCluster(ctx, client, cluster, ControlPlaneMachines(cluster.Name)).
func ControlPlaneMachines(clusterName string) func(machine *kubekeyv1.Machine) bool {
	selector := ControlPlaneSelectorForCluster(clusterName)
	return func(machine *kubekeyv1.Machine) bool {
		if machine == nil {
			return false
		}
		return selector.Matches(labels.Set(machine.Labels))
	}
}

// AdoptableControlPlaneMachines returns a filter to find all un-controlled control plane machines.
// Usage: GetFilteredMachinesForCluster(ctx, client, cluster, AdoptableControlPlaneMachines(cluster.Name, controlPlane)).
func AdoptableControlPlaneMachines(clusterName string) func(machine *kubekeyv1.Machine) bool {
	return And(
		ControlPlaneMachines(clusterName),
		Not(HasControllerRef),
	)
}

// ActiveMachines returns a filter to find all active machines.
// Usage: GetFilteredMachinesForCluster(ctx, client, cluster, ActiveMachines).
func ActiveMachines(machine *kubekeyv1.Machine) bool {
	if machine == nil {
		return false
	}
	return machine.DeletionTimestamp.IsZero()
}

// HasDeletionTimestamp returns a filter to find all machines that have a deletion timestamp.
func HasDeletionTimestamp(machine *kubekeyv1.Machine) bool {
	if machine == nil {
		return false
	}
	return !machine.DeletionTimestamp.IsZero()
}

// IsReady returns a filter to find all machines with the ReadyCondition equals to True.
func IsReady() Func {
	return func(machine *kubekeyv1.Machine) bool {
		if machine == nil {
			return false
		}
		return conditions.IsTrue(machine, kubekeyv1.ReadyCondition)
	}
}

// ShouldRolloutAfter returns a filter to find all machines where
// CreationTimestamp < rolloutAfter < reconciliationTIme.
func ShouldRolloutAfter(reconciliationTime, rolloutAfter *metav1.Time) Func {
	return func(machine *kubekeyv1.Machine) bool {
		if machine == nil {
			return false
		}
		return machine.CreationTimestamp.Before(rolloutAfter) && rolloutAfter.Before(reconciliationTime)
	}
}

// HasAnnotationKey returns a filter to find all machines that have the
// specified Annotation key present.
func HasAnnotationKey(key string) Func {
	return func(machine *kubekeyv1.Machine) bool {
		if machine == nil || machine.Annotations == nil {
			return false
		}
		if _, ok := machine.Annotations[key]; ok {
			return true
		}
		return false
	}
}

// ControlPlaneSelectorForCluster returns the label selector necessary to get control plane machines for a given cluster.
func ControlPlaneSelectorForCluster(clusterName string) labels.Selector {
	must := func(r *labels.Requirement, err error) labels.Requirement {
		if err != nil {
			panic(err)
		}
		return *r
	}
	return labels.NewSelector().Add(
		must(labels.NewRequirement(kubekeyv1.ClusterLabelName, selection.Equals, []string{clusterName})),
		must(labels.NewRequirement(kubekeyv1.MachineControlPlaneLabelName, selection.Exists, []string{})),
	)
}
