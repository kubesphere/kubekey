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
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/storage/names"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/controllers/external"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/conditions"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"

	infrabootstrapv1 "github.com/kubesphere/kubekey/bootstrap/k3s/api/v1beta1"
	infracontrolplanev1 "github.com/kubesphere/kubekey/controlplane/k3s/api/v1beta1"
	k3sCluster "github.com/kubesphere/kubekey/controlplane/k3s/pkg/cluster"
	"github.com/kubesphere/kubekey/controlplane/k3s/pkg/kubeconfig"
	"github.com/kubesphere/kubekey/util/secret"
)

func (r *K3sControlPlaneReconciler) reconcileKubeconfig(ctx context.Context, cluster *clusterv1.Cluster, kcp *infracontrolplanev1.K3sControlPlane) (ctrl.Result, error) {
	endpoint := cluster.Spec.ControlPlaneEndpoint
	if endpoint.IsZero() {
		return ctrl.Result{}, nil
	}

	controllerOwnerRef := *metav1.NewControllerRef(kcp, infracontrolplanev1.GroupVersion.WithKind("K3sControlPlane"))
	clusterName := util.ObjectKey(cluster)
	configSecret, err := secret.GetFromNamespacedName(ctx, r.Client, clusterName, secret.Kubeconfig)
	switch {
	case apierrors.IsNotFound(err):
		createErr := kubeconfig.CreateSecretWithOwner(
			ctx,
			r.Client,
			clusterName,
			endpoint.String(),
			controllerOwnerRef,
		)
		if errors.Is(createErr, kubeconfig.ErrDependentCertificateNotFound) {
			return ctrl.Result{RequeueAfter: dependentCertRequeueAfter}, nil
		}
		// always return if we have just created in order to skip rotation checks
		return ctrl.Result{}, createErr
	case err != nil:
		return ctrl.Result{}, errors.Wrap(err, "failed to retrieve kubeconfig Secret")
	}

	// check if the kubeconfig secret was created by controllers, and thus it has the Cluster as the owner instead of KCP;
	// if yes, adopt it.
	if util.IsOwnedByObject(configSecret, cluster) && !util.IsControlledBy(configSecret, kcp) {
		if err := r.adoptKubeconfigSecret(ctx, cluster, configSecret, controllerOwnerRef); err != nil {
			return ctrl.Result{}, err
		}
	}

	// only do rotation on owned secrets
	if !util.IsControlledBy(configSecret, kcp) {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *K3sControlPlaneReconciler) adoptKubeconfigSecret(ctx context.Context, cluster *clusterv1.Cluster, configSecret *corev1.Secret, controllerOwnerRef metav1.OwnerReference) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Adopting KubeConfig secret created by controllers", "Name", configSecret.Name)

	patch, err := patch.NewHelper(configSecret, r.Client)
	if err != nil {
		return errors.Wrap(err, "failed to create patch helper for the kubeconfig secret")
	}
	configSecret.OwnerReferences = util.RemoveOwnerRef(configSecret.OwnerReferences, metav1.OwnerReference{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "Cluster",
		Name:       cluster.Name,
		UID:        cluster.UID,
	})
	configSecret.OwnerReferences = util.EnsureOwnerRef(configSecret.OwnerReferences, controllerOwnerRef)
	if err := patch.Patch(ctx, configSecret); err != nil {
		return errors.Wrap(err, "failed to patch the kubeconfig secret")
	}
	return nil
}

func (r *K3sControlPlaneReconciler) reconcileExternalReference(ctx context.Context, cluster *clusterv1.Cluster, ref *corev1.ObjectReference) error {
	if !strings.HasSuffix(ref.Kind, clusterv1.TemplateSuffix) {
		return nil
	}

	if err := utilconversion.UpdateReferenceAPIContract(ctx, r.Client, r.APIReader, ref); err != nil {
		return err
	}

	obj, err := external.Get(ctx, r.Client, ref, cluster.Namespace)
	if err != nil {
		return err
	}

	// Note: We intentionally do not handle checking for the paused label on an external template reference

	patchHelper, err := patch.NewHelper(obj, r.Client)
	if err != nil {
		return err
	}

	obj.SetOwnerReferences(util.EnsureOwnerRef(obj.GetOwnerReferences(), metav1.OwnerReference{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "Cluster",
		Name:       cluster.Name,
		UID:        cluster.UID,
	}))

	return patchHelper.Patch(ctx, obj)
}

func (r *K3sControlPlaneReconciler) cloneConfigsAndGenerateMachine(ctx context.Context, cluster *clusterv1.Cluster, kcp *infracontrolplanev1.K3sControlPlane, bootstrapSpec *infrabootstrapv1.K3sConfigSpec, failureDomain *string) error {
	var errs []error

	// Since the cloned resource should eventually have a controller ref for the Machine, we create an
	// OwnerReference here without the Controller field set
	infraCloneOwner := &metav1.OwnerReference{
		APIVersion: infracontrolplanev1.GroupVersion.String(),
		Kind:       "K3sControlPlane",
		Name:       kcp.Name,
		UID:        kcp.UID,
	}

	// Clone the infrastructure template
	infraRef, err := external.CloneTemplate(ctx, &external.CloneTemplateInput{
		Client:      r.Client,
		TemplateRef: &kcp.Spec.MachineTemplate.InfrastructureRef,
		Namespace:   kcp.Namespace,
		OwnerRef:    infraCloneOwner,
		ClusterName: cluster.Name,
		Labels:      k3sCluster.ControlPlaneMachineLabelsForCluster(kcp, cluster.Name),
		Annotations: kcp.Spec.MachineTemplate.ObjectMeta.Annotations,
	})
	if err != nil {
		// Safe to return early here since no resources have been created yet.
		conditions.MarkFalse(kcp, infracontrolplanev1.MachinesCreatedCondition, infracontrolplanev1.InfrastructureTemplateCloningFailedReason,
			clusterv1.ConditionSeverityError, err.Error())
		return errors.Wrap(err, "failed to clone infrastructure template")
	}

	// Clone the bootstrap configuration
	bootstrapRef, err := r.generateK3sConfig(ctx, kcp, cluster, bootstrapSpec)
	if err != nil {
		conditions.MarkFalse(kcp, infracontrolplanev1.MachinesCreatedCondition, infracontrolplanev1.BootstrapTemplateCloningFailedReason,
			clusterv1.ConditionSeverityError, err.Error())
		errs = append(errs, errors.Wrap(err, "failed to generate bootstrap config"))
	}

	// Only proceed to generating the Machine if we haven't encountered an error
	if len(errs) == 0 {
		if err := r.generateMachine(ctx, kcp, cluster, infraRef, bootstrapRef, failureDomain); err != nil {
			conditions.MarkFalse(kcp, infracontrolplanev1.MachinesCreatedCondition, infracontrolplanev1.MachineGenerationFailedReason,
				clusterv1.ConditionSeverityError, err.Error())
			errs = append(errs, errors.Wrap(err, "failed to create Machine"))
		}
	}

	// If we encountered any errors, attempt to clean up any dangling resources
	if len(errs) > 0 {
		if err := r.cleanupFromGeneration(ctx, infraRef, bootstrapRef); err != nil {
			errs = append(errs, errors.Wrap(err, "failed to cleanup generated resources"))
		}

		return kerrors.NewAggregate(errs)
	}

	return nil
}

func (r *K3sControlPlaneReconciler) cleanupFromGeneration(ctx context.Context, remoteRefs ...*corev1.ObjectReference) error {
	var errs []error

	for _, ref := range remoteRefs {
		if ref == nil {
			continue
		}
		config := &unstructured.Unstructured{}
		config.SetKind(ref.Kind)
		config.SetAPIVersion(ref.APIVersion)
		config.SetNamespace(ref.Namespace)
		config.SetName(ref.Name)

		if err := r.Client.Delete(ctx, config); err != nil && !apierrors.IsNotFound(err) {
			errs = append(errs, errors.Wrap(err, "failed to cleanup generated resources after error"))
		}
	}

	return kerrors.NewAggregate(errs)
}

func (r *K3sControlPlaneReconciler) generateK3sConfig(ctx context.Context, kcp *infracontrolplanev1.K3sControlPlane, cluster *clusterv1.Cluster, spec *infrabootstrapv1.K3sConfigSpec) (*corev1.ObjectReference, error) {
	// Create an owner reference without a controller reference because the owning controller is the machine controller
	owner := metav1.OwnerReference{
		APIVersion: infracontrolplanev1.GroupVersion.String(),
		Kind:       "K3sControlPlane",
		Name:       kcp.Name,
		UID:        kcp.UID,
	}

	bootstrapConfig := &infrabootstrapv1.K3sConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.SimpleNameGenerator.GenerateName(kcp.Name + "-"),
			Namespace:       kcp.Namespace,
			Labels:          k3sCluster.ControlPlaneMachineLabelsForCluster(kcp, cluster.Name),
			Annotations:     kcp.Spec.MachineTemplate.ObjectMeta.Annotations,
			OwnerReferences: []metav1.OwnerReference{owner},
		},
		Spec: *spec,
	}

	if err := r.Client.Create(ctx, bootstrapConfig); err != nil {
		return nil, errors.Wrap(err, "Failed to create bootstrap configuration")
	}

	bootstrapRef := &corev1.ObjectReference{
		APIVersion: bootstrapv1.GroupVersion.String(),
		Kind:       "K3sConfig",
		Name:       bootstrapConfig.GetName(),
		Namespace:  bootstrapConfig.GetNamespace(),
		UID:        bootstrapConfig.GetUID(),
	}

	return bootstrapRef, nil
}

func (r *K3sControlPlaneReconciler) generateMachine(ctx context.Context, kcp *infracontrolplanev1.K3sControlPlane, cluster *clusterv1.Cluster, infraRef, bootstrapRef *corev1.ObjectReference, failureDomain *string) error {
	machine := &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:        names.SimpleNameGenerator.GenerateName(kcp.Name + "-"),
			Namespace:   kcp.Namespace,
			Labels:      k3sCluster.ControlPlaneMachineLabelsForCluster(kcp, cluster.Name),
			Annotations: map[string]string{},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(kcp, controlplanev1.GroupVersion.WithKind("K3sControlPlane")),
			},
		},
		Spec: clusterv1.MachineSpec{
			ClusterName:       cluster.Name,
			Version:           &kcp.Spec.Version,
			InfrastructureRef: *infraRef,
			Bootstrap: clusterv1.Bootstrap{
				ConfigRef: bootstrapRef,
			},
			FailureDomain:    failureDomain,
			NodeDrainTimeout: kcp.Spec.MachineTemplate.NodeDrainTimeout,
		},
	}
	if kcp.Spec.MachineTemplate.NodeDeletionTimeout != nil {
		machine.Spec.NodeDeletionTimeout = kcp.Spec.MachineTemplate.NodeDeletionTimeout
	}

	// Machine's bootstrap config may be missing ClusterConfiguration if it is not the first machine in the control plane.
	// We store ClusterConfiguration as annotation here to detect any changes in KCP ClusterConfiguration and rollout the machine if any.
	serverConfig, err := json.Marshal(kcp.Spec.K3sConfigSpec.ServerConfiguration)
	if err != nil {
		return errors.Wrap(err, "failed to marshal cluster configuration")
	}

	// Add the annotations from the MachineTemplate.
	// Note: we intentionally don't use the map directly to ensure we don't modify the map in KCP.
	for k, v := range kcp.Spec.MachineTemplate.ObjectMeta.Annotations {
		machine.Annotations[k] = v
	}
	machine.Annotations[infracontrolplanev1.K3sServerConfigurationAnnotation] = string(serverConfig)

	if err := r.Client.Create(ctx, machine); err != nil {
		return errors.Wrap(err, "failed to create machine")
	}
	return nil
}
