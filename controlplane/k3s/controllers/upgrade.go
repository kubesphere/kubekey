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

	"github.com/pkg/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/collections"
	ctrl "sigs.k8s.io/controller-runtime"

	infracontrolplanev1 "github.com/kubesphere/kubekey/v3/controlplane/k3s/api/v1beta1"
	k3sCluster "github.com/kubesphere/kubekey/v3/controlplane/k3s/pkg/cluster"
)

func (r *K3sControlPlaneReconciler) upgradeControlPlane(
	ctx context.Context,
	cluster *clusterv1.Cluster,
	kcp *infracontrolplanev1.K3sControlPlane,
	controlPlane *k3sCluster.ControlPlane,
	machinesRequireUpgrade collections.Machines,
) (ctrl.Result, error) {
	logger := controlPlane.Logger()

	if kcp.Spec.RolloutStrategy == nil || kcp.Spec.RolloutStrategy.RollingUpdate == nil {
		return ctrl.Result{}, errors.New("rolloutStrategy is not set")
	}

	switch kcp.Spec.RolloutStrategy.Type {
	case infracontrolplanev1.RollingUpdateStrategyType:
		// RolloutStrategy is currently defaulted and validated to be RollingUpdate
		// We can ignore MaxUnavailable because we are enforcing health checks before we get here.
		maxNodes := *kcp.Spec.Replicas + int32(kcp.Spec.RolloutStrategy.RollingUpdate.MaxSurge.IntValue())
		if int32(controlPlane.Machines.Len()) < maxNodes {
			// scaleUp ensures that we don't continue scaling up while waiting for Machines to have NodeRefs
			return r.scaleUpControlPlane(ctx, cluster, kcp, controlPlane)
		}
		return r.scaleDownControlPlane(ctx, cluster, kcp, controlPlane, machinesRequireUpgrade)
	default:
		logger.Info("RolloutStrategy type is not set to RollingUpdateStrategyType, unable to determine the strategy for rolling out machines")
		return ctrl.Result{}, nil
	}
}
