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
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/util/collections"
)

func (r *Reconciler) reconcilePausedCluster(ctx context.Context, clusterScope *scope.ClusterScope, paused bool) (ctrl.Result, error) {
	action := "false"
	if paused {
		action = "true"
	}

	clusterScope.Info(fmt.Sprintf("Reconcile KKCluster paused=%s cluster", action))

	kkCluster := clusterScope.KKCluster
	cluster, err := util.GetOwnerCluster(ctx, r.Client, kkCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}

	if cluster.Spec.Paused == paused {
		clusterScope.V(4).Info(fmt.Sprintf("Skipping reconcilePausedCluster because cluster is already paused=%s", action))
		return ctrl.Result{}, nil
	}

	cluster.Spec.Paused = paused
	if err := r.Client.Update(ctx, cluster); err != nil {
		clusterScope.Error(err, fmt.Sprintf("failed to make the cluster paused=%s, will requeue", action))
		return ctrl.Result{RequeueAfter: 15 * time.Second}, err
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileKKInstanceInPlaceUpgrade(ctx context.Context, clusterScope *scope.ClusterScope,
	filters ...collections.Func) (ctrl.Result, error) {
	clusterScope.Info("Reconcile KKCluster KKInstance in-place upgrade")

	kkCluster := clusterScope.KKCluster
	annotation := kkCluster.GetAnnotations()
	newVersion, ok := annotation[infrav1.InPlaceUpgradeVersionAnnotation]
	if !ok {
		clusterScope.V(4).Info("Skipping reconcileKKInstanceAnnotations because kkCluster does not have annotation %s", infrav1.InPlaceUpgradeVersionAnnotation)
		return ctrl.Result{}, nil
	}

	kkInstances, err := collections.GetFilteredKKInstancesForKKCluster(ctx, r.Client, kkCluster, filters...)
	if err != nil {
		clusterScope.Error(err, "failed to get active kkInstances for kkCluster")
		return ctrl.Result{}, err
	}

	if len(kkInstances) == 0 {
		return ctrl.Result{}, errors.Errorf("no kkInstances found for upgrade")
	}

	for _, kkInstance := range kkInstances {
		if _, ok := kkInstance.Annotations[infrav1.InPlaceUpgradeVersionAnnotation]; !ok {
			kkiCopy := kkInstance.DeepCopy()
			kkiCopy.Annotations[infrav1.InPlaceUpgradeVersionAnnotation] = newVersion

			clusterScope.Info("Patch the KKInstance", "KKInstance", kkiCopy.Name, "annotation", kkiCopy.Annotations)
			if err := r.Client.Update(ctx, kkiCopy); err != nil {
				clusterScope.Error(err, "failed to patch annotation for kkInstance %s", kkiCopy.Name)
				conditions.MarkFalse(kkCluster, infrav1.CallKKInstanceInPlaceUpgradeCondition, infrav1.KKInstanceObjectNotUpdatedReason,
					clusterv1.ConditionSeverityWarning, "Failed to update annotation for kkInstance %s", kkiCopy.Name)
				return ctrl.Result{}, err
			}
		}
	}
	conditions.MarkTrue(kkCluster, infrav1.CallKKInstanceInPlaceUpgradeCondition)
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileKKInstanceUpgradeCheck(ctx context.Context, clusterScope *scope.ClusterScope) (ctrl.Result, error) {
	clusterScope.Info("Reconcile KKCluster KKInstance in-place upgrade check")

	kkCluster := clusterScope.KKCluster
	annotation := kkCluster.GetAnnotations()
	if _, ok := annotation[infrav1.InPlaceUpgradeVersionAnnotation]; !ok {
		clusterScope.V(4).Info("Skipping reconcileKKInstanceAnnotations because kkCluster does not have annotation %s", infrav1.InPlaceUpgradeVersionAnnotation)
		return ctrl.Result{}, nil
	}

	kkInstances, err := collections.GetFilteredKKInstancesForKKCluster(ctx, r.Client, kkCluster, collections.ActiveKKInstances,
		collections.OwnedKKInstances(kkCluster))
	if err != nil {
		clusterScope.Error(err, "failed to get active kkInstances for kkCluster")
		return ctrl.Result{}, err
	}

	conditions.MarkFalse(kkCluster, infrav1.AllKKInstancesUpgradeCompletedCondition, infrav1.WaitingForKKInstancesUpgradeReason,
		clusterv1.ConditionSeverityInfo, "Waiting for all kkInstances upgrade completed")
	for _, kkInstance := range kkInstances {
		if result, err := r.upgradeChecks(ctx, clusterScope, kkInstance); !result.IsZero() || err != nil {
			return result, err
		}
	}
	conditions.MarkTrue(kkCluster, infrav1.AllKKInstancesUpgradeCompletedCondition)
	return ctrl.Result{}, nil
}

func (r *Reconciler) upgradeChecks(_ context.Context, clusterScope *scope.ClusterScope, kkInstance *infrav1.KKInstance) (ctrl.Result, error) {
	clusterScope.V(4).Info("Reconcile KKCluster upgrade checks")

	newVersion, ok := kkInstance.GetAnnotations()[infrav1.InPlaceUpgradeVersionAnnotation]
	if !ok {
		return ctrl.Result{RequeueAfter: upgradeCheckFailedRequeueAfter},
			errors.Errorf("kkInstance %s does not have annotation %s", kkInstance.Name, infrav1.InPlaceUpgradeVersionAnnotation)
	}

	allKKInstanceUpgradeConditions := []clusterv1.ConditionType{
		infrav1.KKInstanceInPlaceUpgradedCondition,
	}

	var kkInstanceErrors []error
	for _, condition := range allKKInstanceUpgradeConditions {
		if err := upgradeChecksCondition("kkinstance", kkInstance, condition); err != nil {
			kkInstanceErrors = append(kkInstanceErrors, err)
		}
		if newVersion != kkInstance.Status.NodeInfo.KubeletVersion {
			kkInstanceErrors = append(kkInstanceErrors, errors.Errorf("kkInstance %s is still upgrading", kkInstance.Name))
		}
	}
	if len(kkInstanceErrors) > 0 {
		aggregatedError := kerrors.NewAggregate(kkInstanceErrors)
		r.Recorder.Eventf(clusterScope.KKCluster, corev1.EventTypeWarning, "KKInstanceInPlaceUpgradeUnFinished",
			"Waiting for the KKInstance to pass upgrade checks to continue reconciliation: %v", aggregatedError)
		clusterScope.Info("Waiting for KKInstance to pass upgrade checks", "failures", aggregatedError.Error())

		return ctrl.Result{RequeueAfter: upgradeCheckFailedRequeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

func upgradeChecksCondition(kind string, obj conditions.Getter, condition clusterv1.ConditionType) error {
	c := conditions.Get(obj, condition)
	if c == nil {
		return errors.Errorf("%s %s does not have %s condition", kind, obj.GetName(), condition)
	}
	if c.Status == corev1.ConditionFalse {
		return errors.Errorf("%s %s reports %s condition is false (%s, %s)", kind, obj.GetName(), condition, c.Severity, c.Message)
	}
	if c.Status == corev1.ConditionUnknown {
		return errors.Errorf("%s %s reports %s condition is unknown (%s)", kind, obj.GetName(), condition, c.Message)
	}
	return nil
}
