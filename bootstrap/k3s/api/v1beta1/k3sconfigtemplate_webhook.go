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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (r *K3sConfigTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/mutate-bootstrap-cluster-x-k8s-io-v1beta1-k3sconfigtemplate,mutating=true,failurePolicy=fail,groups=bootstrap.cluster.x-k8s.io,resources=k3sconfigtemplates,versions=v1beta1,name=default.k3sconfigtemplate.bootstrap.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Defaulter = &K3sConfigTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *K3sConfigTemplate) Default() {
	DefaultK3sConfigSpec(&r.Spec.Template.Spec)
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-bootstrap-cluster-x-k8s-io-v1beta1-k3sconfigtemplate,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=bootstrap.cluster.x-k8s.io,resources=k3sconfigtemplates,versions=v1beta1,name=validation.k3sconfigtemplate.bootstrap.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Validator = &K3sConfigTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *K3sConfigTemplate) ValidateCreate() (admission.Warnings, error) {
	return r.Spec.validate(r.Name)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *K3sConfigTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return r.Spec.validate(r.Name)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *K3sConfigTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (r *K3sConfigTemplateSpec) validate(name string) (admission.Warnings, error) {
	var allErrs field.ErrorList

	allErrs = append(allErrs, r.Template.Spec.Validate(field.NewPath("spec", "template", "spec"))...)

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(GroupVersion.WithKind("K3sConfigTemplate").GroupKind(), name, allErrs)
}
