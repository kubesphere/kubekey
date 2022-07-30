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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage/names"
	capierrors "sigs.k8s.io/cluster-api/errors"
	capiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
)

func (r *KKMachineReconciler) createInstance(
	ctx context.Context,
	machineScope *scope.MachineScope,
	kkInstanceScope scope.KKInstanceScope,
) (*infrav1.KKInstance, error) {

	log := ctrl.LoggerFrom(ctx)
	log.V(4).Info("Creating KKInstance")

	if machineScope.Machine.Spec.Version == nil {
		err := errors.New("Machine's spec.version must be defined")
		machineScope.SetFailureReason(capierrors.CreateMachineError)
		machineScope.SetFailureMessage(err)
		return nil, err
	}

	instanceSpec, err := r.getUnassignedInstanceSpec(machineScope, kkInstanceScope)
	if err != nil {
		return nil, err
	}

	if instanceSpec.Arch == "" {
		instanceSpec.Arch = "amd64"
	}

	// todo: if it need to append a random suffix to the name string
	instanceID := instanceSpec.Name

	gv := infrav1.GroupVersion
	instance := &infrav1.KKInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:            instanceID,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(machineScope.KKMachine, kkMachineKind)},
			Namespace:       machineScope.KKMachine.Namespace,
			// todo: if need to use the kkmachine labels?
			Labels:      machineScope.Machine.Labels,
			Annotations: machineScope.KKMachine.Annotations,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       gv.WithKind("KKInstance").Kind,
			APIVersion: gv.String(),
		},
		Spec: *instanceSpec,
	}

	instance.OwnerReferences = capiutil.EnsureOwnerRef(instance.OwnerReferences, metav1.OwnerReference{
		APIVersion: infrav1.GroupVersion.String(),
		Kind:       "KKCluster",
		Name:       machineScope.InfraCluster.InfraClusterName(),
		UID:        machineScope.InfraCluster.KKCluster.UID,
	})

	if err := r.Client.Create(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

func (r *KKMachineReconciler) generateInstanceID(instanceSpec *infrav1.KKInstanceSpec) string {
	return names.SimpleNameGenerator.GenerateName(instanceSpec.Name + "-")
}

func (r *KKMachineReconciler) getUnassignedInstanceSpec(machineScope *scope.MachineScope, kkInstanceScope scope.KKInstanceScope) (*infrav1.KKInstanceSpec, error) {
	var instanceSpecs []infrav1.KKInstanceSpec
	if machineScope.IsRole(infrav1.ControlPlane) {
		instanceSpecs = kkInstanceScope.GetInstancesSpecByRole(infrav1.ControlPlane)
	} else if machineScope.IsRole(infrav1.Worker) {
		instanceSpecs = kkInstanceScope.GetInstancesSpecByRole(infrav1.Worker)
	} else {
		instanceSpecs = kkInstanceScope.AllInstancesSpec()
	}

	// get all existing instances
	instances, err := kkInstanceScope.AllInstances()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get all existing instance")
	}
	instancesMap := make(map[string]struct{}, 0)
	for _, v := range instances {
		instancesMap[v.Spec.InternalAddress] = struct{}{}
	}

	for _, spec := range instanceSpecs {
		if _, ok := instancesMap[spec.InternalAddress]; !ok {
			auth := kkInstanceScope.GlobalAuth().DeepCopy()
			if err := mergo.Merge(&spec.Auth, auth); err != nil {
				return nil, err
			}
			cm := kkInstanceScope.GlobalContainerManager().DeepCopy()
			if err := mergo.Merge(&spec.ContainerManager, cm); err != nil {
				return nil, err
			}

			return &spec, nil
		}
	}
	return nil, errors.New("unassigned instance not found")
}

func (r *KKMachineReconciler) deleteInstance(ctx context.Context, instance *infrav1.KKInstance) error {
	if err := r.Client.Delete(ctx, instance); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (r *KKMachineReconciler) SetNodeProviderID(ctx context.Context, machineScope *scope.MachineScope, instance *infrav1.KKInstance) error {
	// Usually a cloud provider will do this, but there is no kubekey-cloud provider.
	// Requeue if there is an error, as this is likely momentary load balancer
	// state changes during control plane provisioning.
	remoteClient, err := r.Tracker.GetClient(ctx, client.ObjectKeyFromObject(machineScope.Cluster))
	if err != nil {
		return errors.Wrap(err, "failed to generate workload cluster client")
	}

	node := &corev1.Node{}
	if err = remoteClient.Get(ctx, apimachinerytypes.NamespacedName{Name: instance.Name}, node); err != nil {
		return errors.Wrap(err, "failed to retrieve node")
	}

	machineScope.Info("Setting Kubernetes node providerID")

	patchHelper, err := patch.NewHelper(node, remoteClient)
	if err != nil {
		return err
	}

	node.Spec.ProviderID = machineScope.GetProviderID()

	if err = patchHelper.Patch(ctx, node); err != nil {
		return errors.Wrap(err, "failed update providerID")
	}

	return nil
}
