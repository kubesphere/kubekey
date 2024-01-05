/*
Copyright 2022.

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
	"fmt"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/feature"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	infrabootstrapv1 "github.com/kubesphere/kubekey/v3/bootstrap/k3s/api/v1beta1"
)

const k3sControlPlaneTemplateImmutableMsg = "K3sControlPlaneTemplate spec.template.spec field is immutable. Please create new resource instead."

func (r *K3sControlPlaneTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/mutate-controlplane-cluster-x-k8s-io-v1beta1-k3scontrolplanetemplate,mutating=true,failurePolicy=fail,groups=controlplane.cluster.x-k8s.io,resources=k3scontrolplanetemplates,versions=v1beta1,name=default.k3scontrolplanetemplate.controlplane.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Defaulter = &K3sControlPlaneTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *K3sControlPlaneTemplate) Default() {
	infrabootstrapv1.DefaultK3sConfigSpec(&r.Spec.Template.Spec.K3sConfigSpec)

	r.Spec.Template.Spec.RolloutStrategy = defaultRolloutStrategy(r.Spec.Template.Spec.RolloutStrategy)
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-controlplane-cluster-x-k8s-io-v1beta1-k3scontrolplanetemplate,mutating=false,failurePolicy=fail,groups=controlplane.cluster.x-k8s.io,resources=k3scontrolplanetemplates,versions=v1beta1,name=validation.k3scontrolplanetemplate.controlplane.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Validator = &K3sControlPlaneTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *K3sControlPlaneTemplate) ValidateCreate() (admission.Warnings, error) {
	// NOTE: K3sControlPlaneTemplate is behind ClusterTopology feature gate flag; the web hook
	// must prevent creating new objects in case the feature flag is disabled.
	if !feature.Gates.Enabled(feature.ClusterTopology) {
		return nil, field.Forbidden(
			field.NewPath("spec"),
			"can be set only if the ClusterTopology feature flag is enabled",
		)
	}

	spec := r.Spec.Template.Spec
	allErrs := validateK3sControlPlaneTemplateResourceSpec(spec, field.NewPath("spec", "template", "spec"))
	allErrs = append(allErrs, validateServerConfiguration(spec.K3sConfigSpec.ServerConfiguration, nil, field.NewPath("spec", "template", "spec", "k3sConfigSpec", "serverConfiguration"))...)
	allErrs = append(allErrs, spec.K3sConfigSpec.Validate(field.NewPath("spec", "template", "spec", "k3sConfigSpec"))...)
	if len(allErrs) > 0 {
		return nil, apierrors.NewInvalid(GroupVersion.WithKind("K3sControlPlaneTemplate").GroupKind(), r.Name, allErrs)
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *K3sControlPlaneTemplate) ValidateUpdate(oldRaw runtime.Object) (admission.Warnings, error) {
	var allErrs field.ErrorList
	old, ok := oldRaw.(*K3sControlPlaneTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a K3sControlPlaneTemplate but got a %T", oldRaw))
	}

	if !reflect.DeepEqual(r.Spec.Template.Spec, old.Spec.Template.Spec) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec", "template", "spec"), r, k3sControlPlaneTemplateImmutableMsg),
		)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(GroupVersion.WithKind("K3sControlPlaneTemplate").GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *K3sControlPlaneTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

// validateK3sControlPlaneTemplateResourceSpec is a copy of validateK3sControlPlaneSpec which
// only validates the fields in K3sControlPlaneTemplateResourceSpec we care about.
func validateK3sControlPlaneTemplateResourceSpec(s K3sControlPlaneTemplateResourceSpec, pathPrefix *field.Path) field.ErrorList {
	return validateRolloutStrategy(s.RolloutStrategy, nil, pathPrefix.Child("rolloutStrategy"))
}
