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

package controllers

import (
	"context"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/storage/names"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
)

func (r *KKMachineReconciler) createInstance(ctx context.Context, machineScope *scope.MachineScope, clusterScope pkg.ClusterScoper) (*infrav1.KKInstance, error) {
	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Creating KKInstance")

	instanceSpec, err := r.getUnassignedInstanceSpec(machineScope, clusterScope)
	if err != nil {
		return nil, err
	}

	gv := infrav1.GroupVersion
	instance := &infrav1.KKInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:            r.generateInstanceID(instanceSpec),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(machineScope.KKMachine, kkMachineKind)},
			Namespace:       machineScope.KKMachine.Namespace,
			Labels:          machineScope.KKMachine.Labels,
			Annotations:     machineScope.KKMachine.Annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("KKInstance").Kind,
			APIVersion: gv.String(),
		},
		Spec: *instanceSpec,
	}

	if err := r.Client.Create(ctx, instance); err != nil {
		return nil, err
	}

	// wait until instance running
	log.V(4)
	if err := wait.PollImmediate(r.WaitKKInstanceInterval, r.WaitKKInstanceTimeout, func() (done bool, err error) {

		i := &infrav1.KKInstance{}
		key := client.ObjectKeyFromObject(instance)
		if err := r.Client.Get(ctx, key, i); err != nil {
			return false, err
		}

		if i.Status.State == infrav1.InstanceStateRunning {
			instance = i
			return true, nil
		}
		return false, nil
	}); err != nil {
		return nil, errors.Wrapf(err, "Could not determine if KKInstance is bootstrapped and running.")
	}

	return instance, nil
}

func (r *KKMachineReconciler) generateInstanceID(instanceSpec *infrav1.KKInstanceSpec) string {
	return names.SimpleNameGenerator.GenerateName(instanceSpec.Name + "-")
}

func (r *KKMachineReconciler) getUnassignedInstanceSpec(machineScope *scope.MachineScope, clusterScope pkg.ClusterScoper) (*infrav1.KKInstanceSpec, error) {
	var instanceSpecs []infrav1.KKInstanceSpec
	if machineScope.IsRole(infrav1.ControlPlane) {
		instanceSpecs = clusterScope.GetInstancesSpecByRole(infrav1.ControlPlane)
	} else if machineScope.IsRole(infrav1.Worker) {
		instanceSpecs = clusterScope.GetInstancesSpecByRole(infrav1.Worker)
	} else {
		instanceSpecs = clusterScope.AllInstancesSpec()
	}

	// get all existing instances
	instances, err := clusterScope.AllInstances()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get all existing instance")
	}
	instancesMap := make(map[string]struct{}, 0)
	for _, v := range instances {
		instancesMap[v.Spec.InternalAddress] = struct{}{}
	}

	for _, spec := range instanceSpecs {
		if _, ok := instancesMap[spec.InternalAddress]; !ok {
			auth := clusterScope.Auth().DeepCopy()
			if err := mergo.Merge(&spec.Auth, auth); err != nil {
				return nil, err
			}
			cm := clusterScope.Auth().DeepCopy()
			if err := mergo.Merge(&spec.ContainerManager, cm); err != nil {
				return nil, err
			}

			spec.Bootstrap = machineScope.Machine.Spec.Bootstrap

			return &spec, nil
		}
	}
	return nil, errors.New("unassigned instance not found")
}

func (r *KKMachineReconciler) deleteInstance(ctx context.Context, instance *infrav1.KKInstance) error {
	if err := wait.PollImmediate(r.WaitKKInstanceInterval, r.WaitKKInstanceTimeout, func() (done bool, err error) {
		if err := r.Client.Delete(ctx, instance); err != nil {
			if !apierrors.IsNotFound(err) {
				return false, err
			}
		}

		return true, nil
	}); err != nil {
		return err
	}
	return nil
}
