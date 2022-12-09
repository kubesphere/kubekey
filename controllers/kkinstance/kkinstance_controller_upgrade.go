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
	"fmt"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kubedrain "k8s.io/kubectl/pkg/drain"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	"sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/cluster-api/util"
	capicollections "sigs.k8s.io/cluster-api/util/collections"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/scope"
)

func (r *Reconciler) reconcileInPlaceBinaryService(_ context.Context, instanceScope *scope.InstanceScope, kkInstanceScope scope.KKInstanceScope) (ctrl.Result, error) {
	instanceScope.Info("Reconcile in-place upgrade binary service")
	instanceScope.SetState(infrav1.InstanceStateInPlaceUpgrading)

	kkInstance := instanceScope.KKInstance
	if conditions.IsTrue(kkInstance, infrav1.KKInstanceInPlaceUpgradeBinariesCondition) ||
		instanceScope.InPlaceUpgradeVersion() == kkInstance.Status.NodeInfo.KubeletVersion {
		instanceScope.Info("Skipping reconcileInPlaceBinaryServiceNode")
		return ctrl.Result{}, nil
	}

	clusterScope := instanceScope.InfraCluster
	svc := r.getBinaryService(r.getSSHClient(instanceScope), kkInstanceScope, instanceScope, clusterScope.Distribution())
	if err := svc.UpgradeDownload(r.WaitKKInstanceTimeout); err != nil {
		conditions.MarkFalse(instanceScope.KKInstance, infrav1.KKInstanceInPlaceUpgradeBinariesCondition,
			infrav1.KKInstanceInPlaceGetBinaryFailedReason, clusterv1.ConditionSeverityError, err.Error())
		return ctrl.Result{}, err
	}
	conditions.MarkTrue(instanceScope.KKInstance, infrav1.KKInstanceInPlaceUpgradeBinariesCondition)
	r.Recorder.Event(instanceScope.KKInstance, corev1.EventTypeNormal, "InPlaceUpgradeDownloadBinaries", "In-place upgrade download binaries")
	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileInPlaceKubeadmUpgrade(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	if instanceScope.Machine.Status.NodeRef == nil {
		return ctrl.Result{RequeueAfter: defaultRequeueWait}, errors.Errorf("cannot upgrade machine %s/%s: nodeRef not found", instanceScope.Machine.Namespace, instanceScope.Machine.Name)
	}

	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))
	log.Info("Reconcile in-place kubeadm upgrade")

	kkInstance := instanceScope.KKInstance
	if instanceScope.InPlaceUpgradeVersion() == kkInstance.Status.NodeInfo.KubeletVersion {
		log.V(4).Info("Skipping reconcileInPlaceKubeadmUpgrade, because node kubelet version is already the desired version")
		instanceScope.SetState(infrav1.InstanceStateRunning)
		conditions.MarkTrue(instanceScope.KKInstance, infrav1.KKInstanceInPlaceUpgradedCondition)
		return ctrl.Result{}, nil
	}

	cluster := instanceScope.Cluster
	if res, err := r.kubeadmUpgrade(ctx, instanceScope); !res.IsZero() || err != nil {
		return res, err
	}

	// reconcile drain control plane node
	if r.isNodeDrainAllowed(instanceScope) {
		log.Info("Draining node")

		m := instanceScope.Machine
		if conditions.Get(kkInstance, clusterv1.DrainingSucceededCondition) == nil {
			conditions.MarkFalse(kkInstance, clusterv1.DrainingSucceededCondition, clusterv1.DrainingReason, clusterv1.ConditionSeverityInfo, "Draining the node before in-place upgrade")
		}

		if result, err := r.drainNode(ctx, instanceScope); !result.IsZero() || err != nil {
			if err != nil {
				conditions.MarkFalse(m, clusterv1.DrainingSucceededCondition, clusterv1.DrainingFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
				r.Recorder.Eventf(m, corev1.EventTypeWarning, "FailedDrainNode", "error draining Machine's node %q: %v", m.Status.NodeRef.Name, err)
			}
			return result, err
		}

		conditions.MarkTrue(kkInstance, clusterv1.DrainingSucceededCondition)
		r.Recorder.Eventf(kkInstance, corev1.EventTypeNormal, "SuccessfulDrainNode", "success draining KKInstance's node %q", m.Status.NodeRef.Name)
	}

	// reconcile restart control plane kubelet
	if err := r.restartKubelet(ctx, instanceScope); err != nil {
		log.Error(err, "Failed to restart kubelet")
		r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedInPlaceRestartingKubelet",
			"Failed to restart the kubelet for cluster %s/%s node: %v", cluster.Namespace, cluster.Name, err)
	}

	// reconcile uncordon control plane
	if r.isNodeDrainAllowed(instanceScope) {
		if err := r.cordonOrUncordonNode(ctx, instanceScope, false); err != nil {
			log.Error(err, "Failed to uncordon node")
			r.Recorder.Eventf(instanceScope.KKInstance, corev1.EventTypeWarning, "FailedUncordonNode",
				"Failed to uncordon node for cluster %s/%s node: %v", cluster.Namespace, cluster.Name, err)
		}
	}

	return ctrl.Result{RequeueAfter: defaultRequeueWait}, nil
}

func (r *Reconciler) kubeadmUpgrade(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))

	cluster := instanceScope.Cluster
	ownedMachine, err := instanceScope.InfraCluster.GetMachines(ctx)
	if err != nil {
		log.Error(err, "failed to retrieve owned machines for cluster")
		return ctrl.Result{}, err
	}

	desiredCPMMachine := ownedMachine.Filter(capicollections.ControlPlaneMachines(cluster.Name))
	desiredCPMReplicas := len(desiredCPMMachine)
	numMachines := len(desiredCPMMachine.Filter(capicollections.MatchesKubernetesVersion(instanceScope.InPlaceUpgradeVersion())))
	log.V(4).Info("Number of control plane", "desiredReplicas", desiredCPMReplicas, "numMachines", numMachines)
	switch {
	case instanceScope.IsControlPlane():
		switch {
		case numMachines < desiredCPMReplicas && numMachines == 0:
			log.Info("In-place upgrading the first control plane", "Desired", desiredCPMReplicas, "Existing", numMachines)
			return r.inPlaceUpgradeControlPlane(ctx, instanceScope)
		case numMachines < desiredCPMReplicas && numMachines > 0:
			log.Info("In-place upgrading other control plane", "Desired", desiredCPMReplicas, "Existing", numMachines)
			return r.inPlaceUpgradeScaleUp(ctx, instanceScope)
		}
	case !instanceScope.IsControlPlane():
		if numMachines != desiredCPMReplicas {
			log.Info("Waiting for control plane to be upgraded", "Desired", desiredCPMReplicas, "Existing", numMachines)
			return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
		}

		desiredWorkerMachine := ownedMachine.Filter(capicollections.Not(capicollections.ControlPlaneMachines(cluster.Name)))
		desiredWorkerReplicas := len(desiredWorkerMachine)
		numWorker := len(desiredWorkerMachine.Filter(capicollections.MatchesKubernetesVersion(instanceScope.InPlaceUpgradeVersion())))
		log.V(4).Info("Number of worker", "desiredReplicas", desiredWorkerReplicas, "numMachines", numWorker)

		if numWorker < desiredWorkerReplicas {
			log.Info("In-place upgrading worker nodes", "Desired", desiredWorkerReplicas, "Existing", numWorker)
			if res, err := r.inPlaceUpgradeScaleUpWorker(ctx, instanceScope); !res.IsZero() || err != nil {
				if err != nil {
					log.Error(err, "Failed to in-place upgrade worker")
				}
				return res, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) isNodeDrainAllowed(instanceScope *scope.InstanceScope) bool {
	if _, exist := instanceScope.Machine.GetAnnotations()[clusterv1.ExcludeNodeDrainingAnnotation]; exist {
		return false
	}
	if r.nodeDrainTimeoutExceeded(instanceScope.Machine) {
		return false
	}
	return true
}

func (r *Reconciler) nodeDrainTimeoutExceeded(machine *clusterv1.Machine) bool {
	// if the NodeDrainTimeout type is not set by user
	if machine.Spec.NodeDrainTimeout == nil || machine.Spec.NodeDrainTimeout.Seconds() <= 0 {
		return false
	}

	// if the draining succeeded condition does not exist
	if conditions.Get(machine, clusterv1.DrainingSucceededCondition) == nil {
		return false
	}

	now := time.Now()
	firstTimeDrain := conditions.GetLastTransitionTime(machine, clusterv1.DrainingSucceededCondition)
	diff := now.Sub(firstTimeDrain.Time)
	return diff.Seconds() >= machine.Spec.NodeDrainTimeout.Seconds()
}

func (r *Reconciler) drainNode(ctx context.Context, instanceScope *scope.InstanceScope) (ctrl.Result, error) {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))

	cluster := instanceScope.Cluster
	restConfig, err := remote.RESTConfig(ctx, controllerName, r.Client, util.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error creating a remote client while deleting Machine, won't retry")
		return ctrl.Result{}, nil
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Error(err, "Error creating a remote client while deleting Machine, won't retry")
		return ctrl.Result{}, nil
	}

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If an admin deletes the node directly, we'll end up here.
			log.Error(err, "Could not find node from noderef, it may have already been upgrading")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrapf(err, "unable to get node %v", nodeName)
	}

	drainer := &kubedrain.Helper{
		Client:              kubeClient,
		Ctx:                 ctx,
		Force:               true,
		IgnoreAllDaemonSets: true,
		DeleteEmptyDirData:  true,
		GracePeriodSeconds:  -1,
		// If a pod is not evicted in 20 seconds, retry the eviction next time the
		// machine gets reconciled again (to allow other machines to be reconciled).
		Timeout: 20 * time.Second,
		OnPodDeletedOrEvicted: func(pod *corev1.Pod, usingEviction bool) {
			verbStr := "Deleted"
			if usingEviction {
				verbStr = "Evicted"
			}
			log.Info(fmt.Sprintf("%s pod from Node", verbStr),
				"Pod", klog.KObj(pod))
		},
		Out: writer{log.Info},
		ErrOut: writer{func(msg string, keysAndValues ...interface{}) {
			log.Error(nil, msg, keysAndValues...)
		}},
	}

	if noderefutil.IsNodeUnreachable(node) {
		// When the node is unreachable and some pods are not evicted for as long as this timeout, we ignore them.
		drainer.SkipWaitForDeleteTimeoutSeconds = 60 * 5 // 5 minutes
	}

	if err := kubedrain.RunCordonOrUncordon(drainer, node, true); err != nil {
		// Machine will be re-reconciled after a cordon failure.
		log.Error(err, "Cordon failed")
		return ctrl.Result{}, errors.Wrapf(err, "unable to cordon node %v", node.Name)
	}

	if err := kubedrain.RunNodeDrain(drainer, node.Name); err != nil {
		// Machine will be re-reconciled after a drain failure.
		log.Error(err, "Drain failed, retry in 20s")
		return ctrl.Result{RequeueAfter: 20 * time.Second}, nil
	}

	log.Info("Drain successful")
	return ctrl.Result{}, nil
}

func (r *Reconciler) restartKubelet(_ context.Context, instanceScope *scope.InstanceScope) error {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))
	log.Info("In-place restart kubelet")

	if _, err := r.getSSHClient(instanceScope).SudoCmdf("systemctl daemon-reload && systemctl restart kubelet"); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) cordonOrUncordonNode(ctx context.Context, instanceScope *scope.InstanceScope, desired bool) error {
	nodeName := instanceScope.Machine.Status.NodeRef.Name
	action := "cordon"
	if !desired {
		action = "uncordon"
	}

	log := instanceScope.Logger.WithValues("Node", klog.KRef("", nodeName))
	log.Info(fmt.Sprintf("In-place %s node", action))

	cluster := instanceScope.Cluster
	restConfig, err := remote.RESTConfig(ctx, controllerName, r.Client, util.ObjectKey(cluster))
	if err != nil {
		log.Error(err, fmt.Sprintf("Error creating a remote client while %s node, won't retry", action))
		return nil
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error creating a remote client while %s bide, won't retry", action))
		return nil
	}

	node, err := kubeClient.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// If an admin deletes the node directly, we'll end up here.
			log.Error(err, "Could not find node from noderef, it may have already been upgrading")
			return nil
		}
		return errors.Wrapf(err, "unable to get node %v", nodeName)
	}

	drainer := &kubedrain.Helper{
		Client: kubeClient,
		Ctx:    ctx,
	}

	if err := kubedrain.RunCordonOrUncordon(drainer, node, desired); err != nil {
		// Machine will be re-reconciled after a cordon failure.
		log.Error(err, fmt.Sprintf("%s failed", action))
		return errors.Wrapf(err, "unable to %s node %v", action, node.Name)
	}
	return nil
}

// writer implements io.Writer interface as a pass-through for klog.
type writer struct {
	logFunc func(msg string, keysAndValues ...interface{})
}

// Write passes string(p) into writer's logFunc and always returns len(p).
func (w writer) Write(p []byte) (n int, err error) {
	w.logFunc(string(p))
	return len(p), nil
}
