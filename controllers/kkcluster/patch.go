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

package kkcluster

import (
	"context"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	capicollections "sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg/scope"
	"github.com/kubesphere/kubekey/v3/util/collections"
)

func (r *Reconciler) reconcilePatchAnnotations(ctx context.Context, clusterScope *scope.ClusterScope) (res ctrl.Result, err error) {
	clusterScope.Info("Reconcile KKCluster patch KKCluster and KKInstance")

	defer func() {
		if err != nil && res.IsZero() {
			clusterScope.Info("Reconcile KKCluster patch annotations failed, requeue")
			res = ctrl.Result{RequeueAfter: 15 * time.Second}
		}
	}()

	kkCluster := clusterScope.KKCluster
	kkInstances, err := collections.GetFilteredKKInstancesForKKCluster(ctx, r.Client, kkCluster, collections.ActiveKKInstances,
		collections.OwnedKKInstances(kkCluster))
	if err != nil {
		clusterScope.Error(err, "failed to get active kkInstances for kkCluster")
		return ctrl.Result{}, err
	}

	for _, kkInstance := range kkInstances {
		if _, ok := kkInstance.GetAnnotations()[infrav1.InPlaceUpgradeVersionAnnotation]; ok {
			clusterScope.V(4).Info("KKCluster patch annotations for kkInstance",
				"kkInstance", kkInstance.Name)
			// delete in-place upgrade annotation
			kkInstanceAnnotation := kkInstance.GetAnnotations()
			delete(kkInstanceAnnotation, infrav1.InPlaceUpgradeVersionAnnotation)
			kkInstance.SetAnnotations(kkInstanceAnnotation)

			if err := r.Client.Update(ctx, kkInstance); err != nil {
				clusterScope.Error(err, "failed to update kkInstance")
				return ctrl.Result{}, err
			}
		}

		if conditions.Has(kkInstance, infrav1.KKInstanceInPlaceUpgradeBinariesCondition) {
			clusterScope.V(4).Info("KKCluster patch conditions for kkInstance",
				"kkInstance", kkInstance.Name)
			// delete in-place upgrade condition
			conditions.Delete(kkInstance, infrav1.KKInstanceInPlaceUpgradeBinariesCondition)

			if err := r.Client.Status().Update(ctx, kkInstance); err != nil {
				clusterScope.Error(err, "failed to update kkInstance status")
				return ctrl.Result{}, err
			}
		}
	}

	annotation := kkCluster.GetAnnotations()
	if _, ok := annotation[infrav1.InPlaceUpgradeVersionAnnotation]; ok {
		clusterScope.V(4).Info("KKCluster patch annotations")
		delete(annotation, infrav1.InPlaceUpgradeVersionAnnotation)
		kkCluster.SetAnnotations(annotation)
		patchHelper, err := patch.NewHelper(kkCluster, r.Client)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err = patchHelper.Patch(ctx, kkCluster); err != nil {
			clusterScope.Error(err, "failed to patch KKCluster")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcilePatchResourceSpecVersion(ctx context.Context, clusterScope *scope.ClusterScope) (res ctrl.Result, err error) {
	clusterScope.Info("Reconcile KKCluster patch resource spec version")

	defer func() {
		if err != nil && res.IsZero() {
			res = ctrl.Result{RequeueAfter: 15 * time.Second}
		}
	}()

	kkCluster := clusterScope.KKCluster
	annotation := kkCluster.GetAnnotations()
	newVersion, ok := annotation[infrav1.InPlaceUpgradeVersionAnnotation]
	if !ok {
		clusterScope.V(4).Info("Skipping reconcilePatchResourceSpecVersion because kkCluster does not have annotation %s", infrav1.InPlaceUpgradeVersionAnnotation)
		return ctrl.Result{}, nil
	}

	machines, err := clusterScope.GetMachines(ctx, capicollections.MatchesKubernetesVersion(newVersion))
	if err != nil {
		clusterScope.Error(err, "failed to get machines for cluster")
		return ctrl.Result{}, err
	}

	clusterScope.Info("Prepare to patch the .spec.version of the machines' owner", "numMachine", machines.Len())

	if res, err := r.patchKCPSpecVersion(ctx, clusterScope, newVersion, machines); !res.IsZero() || err != nil {
		return res, err
	}

	if res, err := r.patchMachineSet(ctx, clusterScope, newVersion, machines); !res.IsZero() || err != nil {
		return res, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) patchKCPSpecVersion(ctx context.Context, clusterScope *scope.ClusterScope, newVersion string, machines capicollections.Machines) (ctrl.Result, error) {
	cluster := clusterScope.Cluster
	kcpRef := cluster.Spec.ControlPlaneRef

	if kcpRef == nil {
		return ctrl.Result{}, nil
	}

	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: kcpRef.Namespace, Name: kcpRef.Name}, kcp); err != nil {
		if apierrors.IsNotFound(errors.Cause(err)) {
			clusterScope.Info("Could not find KubeadmControlPlane, requeuing", "KubeadmControlPlane", kcpRef.Name)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	// can not match any upgraded machine
	if machines.Filter(capicollections.OwnedMachines(kcp)).Len() == 0 {
		clusterScope.Info("Could not find any upgraded control plane machine", "KubeadmControlPlane", kcp.Name)
		return ctrl.Result{}, nil
	}

	clusterScope.Info("Patching the KCP with the upgraded K8s version", "KubeadmControlPlane", kcp.Name)
	kcp.Spec.Version = newVersion
	if err := r.Client.Update(ctx, kcp); err != nil {
		clusterScope.Error(err, "failed to patch KubeadmControlPlane %s", kcp.Name)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchMachineSet(ctx context.Context, clusterScope *scope.ClusterScope, newVersion string, machines capicollections.Machines) (ctrl.Result, error) {
	cluster := clusterScope.Cluster
	machineSets := &clusterv1.MachineSetList{}

	if err := r.Client.List(ctx, machineSets,
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels{clusterv1.ClusterNameLabel: cluster.Name}); err != nil {
		return ctrl.Result{}, err
	}

	msMap := make(map[string]*clusterv1.MachineSet)
	for i := range machineSets.Items {
		ms := machineSets.Items[i]
		if ms.Spec.ClusterName != cluster.Name {
			continue
		}

		if machines.Filter(capicollections.OwnedMachines(&ms)).Len() == 0 {
			continue
		}

		clusterScope.Info("Patching the machineSet with the upgraded K8s version", "MachineSet", ms.Name)
		ms.Spec.Template.Spec.Version = pointer.String(newVersion)
		if err := r.Client.Update(ctx, &ms); err != nil {
			clusterScope.Error(err, "failed to patch MachineSet %s", ms.Name)
			return ctrl.Result{}, err
		}
		msMap[ms.Name] = &ms
	}

	if res, err := r.patchMachineDeployment(ctx, clusterScope, newVersion, msMap); !res.IsZero() || err != nil {
		if err != nil {
			return res, err
		}
		return res, nil
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) patchMachineDeployment(ctx context.Context, clusterScope *scope.ClusterScope, newVersion string, machinesets map[string]*clusterv1.MachineSet) (ctrl.Result, error) {
	cluster := clusterScope.Cluster
	machineDeployments := &clusterv1.MachineDeploymentList{}

	if err := r.Client.List(ctx, machineDeployments,
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels{clusterv1.ClusterNameLabel: cluster.Name}); err != nil {
		return ctrl.Result{}, err
	}

	for i := range machineDeployments.Items {
		md := machineDeployments.Items[i]
		if md.Spec.ClusterName != cluster.Name {
			continue
		}

		if !hasOwnGivenMachineSet(machinesets, &md) {
			continue
		}

		clusterScope.Info("Patching the machineDeployment with the upgraded K8s version", "MachineDeployment", md.Name)
		md.Spec.Template.Spec.Version = pointer.String(newVersion)
		if err := r.Client.Update(ctx, &md); err != nil {
			clusterScope.Error(err, "failed to patch MachineDeployment %s", md.Name)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func hasOwnGivenMachineSet(machinesets map[string]*clusterv1.MachineSet, owner *clusterv1.MachineDeployment) bool {
	for _, ms := range machinesets {
		if util.IsOwnedByObject(ms, owner) {
			return true
		}
	}
	return false
}
