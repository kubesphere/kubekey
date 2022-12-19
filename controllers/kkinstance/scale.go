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

package kkinstance

import (
	"context"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	capicollections "sigs.k8s.io/cluster-api/util/collections"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kubesphere/kubekey/v3/pkg/scope"
)

func (r *Reconciler) inPlaceUpgradeControlPlane(ctx context.Context, instanceScope *scope.InstanceScope) (_ ctrl.Result, retErr error) {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))

	if !instanceScope.IsControlPlane() {
		return ctrl.Result{}, nil
	}

	cluster := instanceScope.Cluster
	newVersionCPM, err := instanceScope.InfraCluster.GetMachines(ctx, capicollections.ControlPlaneMachines(cluster.Name),
		capicollections.MatchesKubernetesVersion(instanceScope.InPlaceUpgradeVersion()))
	if err != nil {
		log.Error(err, "Failed to perform an uncached read of new version control plane machines")
		return ctrl.Result{}, err
	}

	if len(newVersionCPM) > 0 {
		return ctrl.Result{}, errors.Errorf(
			"already have the first control plane upgrade completed, found %d new version control plane for cluster %s/%s: controller cache or management cluster is misbehaving",
			len(newVersionCPM), cluster.Namespace, cluster.Name,
		)
	}

	// acquire the lock so that only the first machine configured
	// as control plane get processed here
	// if not the first, requeue
	if !r.Lock.Lock(ctx, cluster, instanceScope.KKInstance) {
		log.Info("A control plane is already upgrading, requeueing until the first control plane is upgraded")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	defer func() {
		if retErr != nil {
			if !r.Lock.Unlock(ctx, cluster) {
				retErr = kerrors.NewAggregate([]error{retErr, errors.New("failed to unlock the lock")})
			}
		}
	}()

	if _, err := r.getSSHClient(instanceScope).SudoCmdf("kubeadm upgrade apply %s -y "+
		"--ignore-preflight-errors=all "+
		"--allow-experimental-upgrades "+
		"--allow-release-candidate-upgrades "+
		"--etcd-upgrade=false "+
		"--certificate-renewal=true", instanceScope.InPlaceUpgradeVersion()); err != nil {
		log.Error(err, "Failed to upgrade the first control plane")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to upgrade the first control plane for cluster %s/%s control plane: %v", cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}

	//  reconcile patch machine spec version
	log.V(4).Info("Patching machine spec version")
	if err := r.patchMachineSpecVersion(ctx, instanceScope); err != nil {
		log.Error(err, "Failed to patch machine spec version")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to patch machine %s spec version for cluster %s/%s machine: %v", instanceScope.Machine, cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) inPlaceUpgradeScaleUp(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))

	if instanceScope.InPlaceUpgradeVersion() == *instanceScope.Machine.Spec.Version {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	cluster := instanceScope.Cluster
	r.Lock.Unlock(ctx, cluster)

	if _, err := r.getSSHClient(instanceScope).SudoCmdf("kubeadm upgrade node"); err != nil {
		log.Error(err, "Failed to upgrade the control plane")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to upgrade the node for cluster %s/%s control plane: %v", cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}

	//  reconcile patch machine spec version
	log.V(4).Info("Patching machine spec version")
	if err := r.patchMachineSpecVersion(ctx, instanceScope); err != nil {
		log.Error(err, "Failed to patch machine spec version")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to patch machine %s spec version for cluster %s/%s machine: %v", instanceScope.Machine, cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *Reconciler) inPlaceUpgradeScaleUpWorker(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))

	cluster := instanceScope.Cluster
	if _, err := r.getSSHClient(instanceScope).SudoCmdf("kubeadm upgrade node"); err != nil {
		log.Error(err, "Failed to upgrade the node")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to upgrade the node for cluster %s/%s worker: %v", cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}

	//  reconcile patch machine spec version
	log.V(4).Info("Patching machine spec version")
	if err := r.patchMachineSpecVersion(ctx, instanceScope); err != nil {
		log.Error(err, "Failed to patch machine spec version")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceUpgradeControlPlane",
			"Failed to patch machine %s spec version for cluster %s/%s machine: %v", instanceScope.Machine, cluster.Namespace, cluster.Name, err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) patchMachineSpecVersion(ctx context.Context, instanceScope *scope.InstanceScope) error {
	machineCopy := instanceScope.Machine.DeepCopy()
	machineCopy.Spec.Version = pointer.String(instanceScope.InPlaceUpgradeVersion())
	return r.Client.Update(ctx, machineCopy)
}
