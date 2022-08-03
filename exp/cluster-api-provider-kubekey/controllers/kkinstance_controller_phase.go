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

	ctrl "sigs.k8s.io/controller-runtime"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
)

func (r *KKInstanceReconciler) reconcilePing(ctx context.Context, instanceScope *scope.InstanceScope) error {
	log := ctrl.LoggerFrom(ctx, "infraCluster", instanceScope.InfraCluster.Name())
	log.V(4).Info("Reconcile ping")

	sshClient := r.getSSHClient(instanceScope)
	var err error
	for i := 0; i < 3; i++ {
		err = sshClient.Ping()
		if err == nil {
			break
		}
	}
	return err
}

func (r *KKInstanceReconciler) reconcileBootstrap(ctx context.Context, sshClient ssh.Interface, instanceScope *scope.InstanceScope, lbScope scope.LBScope) error {
	log := ctrl.LoggerFrom(ctx, "infraCluster", instanceScope.InfraCluster.Name())
	log.V(4).Info("Reconcile bootstrap")

	instanceScope.SetState(infrav1.InstanceStateBootstrapping)

	svc := r.getBootstrapService(sshClient, lbScope)

	if err := svc.AddUsers(); err != nil {
		return err
	}
	if err := svc.CreateDirectory(); err != nil {
		return err
	}
	if err := svc.ResetTmpDirectory(); err != nil {
		return err
	}
	if err := svc.ExecInitScript(); err != nil {
		return err
	}
	return nil
}
