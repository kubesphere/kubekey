/*
Copyright 2020 The KubeSphere Authors.

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
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/api/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func installTasks(r *ClusterReconciler, ctx context.Context, cluster *kubekeyv1alpha1.Cluster, mgr *manager.Manager) error {
	// init nodes
	cluster.Status.Conditions = []kubekeyv1alpha1.Condition{}
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}
	initNodesCondition := kubekeyv1alpha1.Condition{
		Step:      "Init nodes",
		StartTime: metav1.Now(),
		EndTime:   metav1.Now(),
		Status:    false,
	}
	cluster.Status.Conditions = append(cluster.Status.Conditions, initNodesCondition)
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	if err := runTasks(mgr, initNodesTasks); err != nil {
		return err
	}

	cluster.Status.Conditions[0].EndTime = metav1.Now()
	cluster.Status.Conditions[0].Status = true
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	// pull images
	pullImagesCondition := kubekeyv1alpha1.Condition{
		Step:      "Pull images",
		StartTime: metav1.Now(),
		EndTime:   metav1.Now(),
		Status:    false,
	}
	cluster.Status.Conditions = append(cluster.Status.Conditions, pullImagesCondition)
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	if err := runTasks(mgr, pullImagesTaks); err != nil {
		return err
	}

	cluster.Status.Conditions[1].EndTime = metav1.Now()
	cluster.Status.Conditions[1].Status = true
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	// init etcd cluster
	initEtcdClusterCondition := kubekeyv1alpha1.Condition{
		Step:      "init etcd cluster",
		StartTime: metav1.Now(),
		EndTime:   metav1.Now(),
		Status:    false,
	}
	cluster.Status.Conditions = append(cluster.Status.Conditions, initEtcdClusterCondition)
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	if err := runTasks(mgr, initEtcdClusterTasks); err != nil {
		return err
	}

	cluster.Status.Conditions[2].EndTime = metav1.Now()
	cluster.Status.Conditions[2].Status = true
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	// init controle plane
	initControlPlaneCondition := kubekeyv1alpha1.Condition{
		Step:      "init control plane",
		StartTime: metav1.Now(),
		EndTime:   metav1.Now(),
		Status:    false,
	}
	cluster.Status.Conditions = append(cluster.Status.Conditions, initControlPlaneCondition)
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	if err := runTasks(mgr, initControlPlaneTasks); err != nil {
		return err
	}

	cluster.Status.Conditions[3].EndTime = metav1.Now()
	cluster.Status.Conditions[3].Status = true
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	// join nodes
	joinNodesCondition := kubekeyv1alpha1.Condition{
		Step:      "join nodes",
		StartTime: metav1.Now(),
		EndTime:   metav1.Now(),
		Status:    false,
	}
	cluster.Status.Conditions = append(cluster.Status.Conditions, joinNodesCondition)
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	if err := runTasks(mgr, joinNodesTasks); err != nil {
		return err
	}

	cluster.Status.Conditions[4].EndTime = metav1.Now()
	cluster.Status.Conditions[4].Status = true
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	cluster.Status.Version = mgr.Cluster.Kubernetes.Version
	cluster.Status.NodesCount = len(mgr.AllNodes)
	cluster.Status.MasterCount = len(mgr.MasterNodes)
	cluster.Status.WorkerCount = len(mgr.WorkerNodes)
	cluster.Status.EtcdCount = len(mgr.EtcdNodes)
	cluster.Status.NetworkPlugin = mgr.Cluster.Network.Plugin

	cluster.Status.Nodes = []kubekeyv1alpha1.NodeStatus{}
	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	for _, node := range mgr.AllNodes {
		cluster.Status.Nodes = append(cluster.Status.Nodes, kubekeyv1alpha1.NodeStatus{
			InternalIP: node.InternalAddress,
			Hostname:   node.Name,
			Roles:      map[string]bool{"etcd": node.IsEtcd, "master": node.IsMaster, "worker": node.IsWorker},
		})
	}

	if err := r.Status().Update(ctx, cluster); err != nil {
		return err
	}

	return nil
}
