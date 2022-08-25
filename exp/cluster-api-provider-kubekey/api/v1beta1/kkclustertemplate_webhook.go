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

package v1beta1

import (
	"github.com/google/go-cmp/cmp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kkclustertemplatelog = logf.Log.WithName("kkclustertemplate-resource")

// SetupWebhookWithManager sets up and registers the webhook with the manager.
func (r *KKClusterTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkclustertemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclustertemplates,verbs=create;update,versions=v1beta1,name=default.kkclustertemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKClusterTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KKClusterTemplate) Default() {
	kkclustertemplatelog.Info("default", "name", r.Name)

	defaultAuth(&r.Spec.Template.Spec.Nodes.Auth)
	defaultContainerManager(&r.Spec.Template.Spec)
	defaultInstance(&r.Spec.Template.Spec)
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkclustertemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclustertemplates,verbs=create;update,versions=v1beta1,name=validation.kkclustertemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKClusterTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KKClusterTemplate) ValidateCreate() error {
	kkclustertemplatelog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateClusterNodes(r.Spec.Template.Spec.Nodes)...)
	allErrs = append(allErrs, validateLoadBalancer(r.Spec.Template.Spec.ControlPlaneLoadBalancer)...)

	return aggregateObjErrors(r.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KKClusterTemplate) ValidateUpdate(old runtime.Object) error {
	kkclustertemplatelog.Info("validate update", "name", r.Name)

	oldC := old.(*KKClusterTemplate)
	if !cmp.Equal(r.Spec, oldC.Spec) {
		return apierrors.NewBadRequest("KKClusterTemplate.Spec is immutable")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KKClusterTemplate) ValidateDelete() error {
	return nil
}
