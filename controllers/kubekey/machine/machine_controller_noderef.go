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

package machine

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha3"
	"github.com/kubesphere/kubekey/util"
	"github.com/kubesphere/kubekey/util/annotations"
	"github.com/kubesphere/kubekey/util/conditions"
	"github.com/kubesphere/kubekey/util/patch"
)

var (
	// ErrNodeNotFound signals that a corev1.Node could not be found for the given node name.
	ErrNodeNotFound = errors.New("cannot find node with matching node name")
)

func (r *MachineReconciler) reconcileNode(ctx context.Context, cluster *kubekeyv1.Cluster, machine *kubekeyv1.Machine) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "machine", machine.Name, "namespace", machine.Namespace)
	log = log.WithValues("cluster", cluster.Name)

	// Check that the Machine has a valid name.
	if machine.Spec.Name == "" {
		log.Info("Cannot reconcile Machine's Node, no valid node name yet")
		conditions.MarkFalse(machine, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.WaitingForNodeRefReason, kubekeyv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	}

	remoteClient, err := r.Tracker.GetClient(ctx, util.ObjectKey(cluster))
	if err != nil {
		return ctrl.Result{}, err
	}

	// Even if Status.NodeRef exists, continue to do the following checks to make sure Node is healthy
	node, err := r.getNode(ctx, remoteClient, machine.Spec.Name, machine.Namespace)
	if err != nil {
		if err == ErrNodeNotFound {
			// While a NodeRef is set in the status, failing to get that node means the node is deleted.
			// If Status.NodeRef is not set before, node still can be in the provisioning state.
			if machine.Status.NodeRef != nil {
				conditions.MarkFalse(machine, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.NodeNotFoundReason, kubekeyv1.ConditionSeverityError, "")
				return ctrl.Result{}, errors.Wrapf(err, "no matching Node for Machine %q in namespace %q", machine.Name, machine.Namespace)
			}
			conditions.MarkFalse(machine, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.NodeProvisioningReason, kubekeyv1.ConditionSeverityWarning, "")
			// No need to requeue here. Nodes emit an event that triggers reconciliation.
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to retrieve Node by ProviderID")
		r.recorder.Event(machine, corev1.EventTypeWarning, "Failed to retrieve Node by node name", err.Error())
		return ctrl.Result{}, err
	}

	// Set the Machine NodeRef.
	if machine.Status.NodeRef == nil {
		machine.Status.NodeRef = &corev1.ObjectReference{
			Kind:       node.Kind,
			APIVersion: node.APIVersion,
			Name:       node.Name,
			UID:        node.UID,
		}
		log.Info("Set Machine's NodeRef", "noderef", machine.Status.NodeRef.Name)
		r.recorder.Event(machine, corev1.EventTypeNormal, "SuccessfulSetNodeRef", machine.Status.NodeRef.Name)
	}

	// Set the NodeSystemInfo.
	machine.Status.NodeInfo = &node.Status.NodeInfo

	// Reconcile node annotations.
	patchHelper, err := patch.NewHelper(node, remoteClient)
	if err != nil {
		return ctrl.Result{}, err
	}
	desired := map[string]string{
		kubekeyv1.ClusterNameAnnotation:      machine.Spec.ClusterName,
		kubekeyv1.ClusterNamespaceAnnotation: machine.GetNamespace(),
		kubekeyv1.MachineAnnotation:          machine.Name,
	}
	if owner := metav1.GetControllerOfNoCopy(machine); owner != nil {
		desired[kubekeyv1.OwnerKindAnnotation] = owner.Kind
		desired[kubekeyv1.OwnerNameAnnotation] = owner.Name
	}
	if annotations.AddAnnotations(node, desired) {
		if err := patchHelper.Patch(ctx, node); err != nil {
			log.V(2).Info("Failed patch node to set annotations", "err", err, "node name", node.Name)
			return ctrl.Result{}, err
		}
	}

	// Do the remaining node health checks, then set the node health to true if all checks pass.
	status, message := summarizeNodeConditions(node)
	if status == corev1.ConditionFalse {
		conditions.MarkFalse(machine, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.NodeConditionsFailedReason, kubekeyv1.ConditionSeverityWarning, message)
		return ctrl.Result{}, nil
	}
	if status == corev1.ConditionUnknown {
		conditions.MarkUnknown(machine, kubekeyv1.MachineNodeHealthyCondition, kubekeyv1.NodeConditionsFailedReason, message)
		return ctrl.Result{}, nil
	}

	conditions.MarkTrue(machine, kubekeyv1.MachineNodeHealthyCondition)
	return ctrl.Result{}, nil
}

// summarizeNodeConditions summarizes a Node's conditions and returns the summary of condition statuses and concatenate failed condition messages:
// if there is at least 1 semantically-negative condition, summarized status = False;
// if there is at least 1 semantically-positive condition when there is 0 semantically negative condition, summarized status = True;
// if all conditions are unknown,  summarized status = Unknown.
// (semantically true conditions: NodeMemoryPressure/NodeDiskPressure/NodePIDPressure == false or Ready == true.)
func summarizeNodeConditions(node *corev1.Node) (corev1.ConditionStatus, string) {
	semanticallyFalseStatus := 0
	unknownStatus := 0

	message := ""
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeMemoryPressure, corev1.NodeDiskPressure, corev1.NodePIDPressure:
			if condition.Status != corev1.ConditionFalse {
				message += fmt.Sprintf("Node condition %s is %s", condition.Type, condition.Status) + ". "
				if condition.Status == corev1.ConditionUnknown {
					unknownStatus++
					continue
				}
				semanticallyFalseStatus++
			}
		case corev1.NodeReady:
			if condition.Status != corev1.ConditionTrue {
				message += fmt.Sprintf("Node condition %s is %s", condition.Type, condition.Status) + ". "
				if condition.Status == corev1.ConditionUnknown {
					unknownStatus++
					continue
				}
				semanticallyFalseStatus++
			}
		}
	}
	if semanticallyFalseStatus > 0 {
		return corev1.ConditionFalse, message
	}
	if semanticallyFalseStatus+unknownStatus < 4 {
		return corev1.ConditionTrue, message
	}
	return corev1.ConditionUnknown, message
}

func (r *MachineReconciler) getNode(ctx context.Context, c client.Reader, name string, namespace string) (*corev1.Node, error) {
	node := corev1.Node{}
	if err := c.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &node); err != nil {
		return nil, err
	}

	return &node, nil
}
