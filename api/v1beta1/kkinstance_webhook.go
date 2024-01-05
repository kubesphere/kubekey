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
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/cluster-api/util/version"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kkinstancelog = logf.Log.WithName("kkinstance-resource")

func (k *KKInstance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkinstance,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkinstances,verbs=create;update,versions=v1beta1,name=default.kkinstance.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKInstance{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *KKInstance) Default() {
	kkinstancelog.Info("default", "name", k.Name)
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkinstances,verbs=create;update,versions=v1beta1,name=validation.kkinstance.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKInstance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KKInstance) ValidateCreate() (admission.Warnings, error) {
	kkinstancelog.Info("validate create", "name", k.Name)
	return admission.Warnings{}, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KKInstance) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	kkinstancelog.Info("validate update", "name", k.Name)

	var allErrs field.ErrorList
	if v, ok := k.GetAnnotations()[InPlaceUpgradeVersionAnnotation]; ok {
		if k.Status.NodeInfo == nil {
			allErrs = append(allErrs,
				field.Invalid(
					field.NewPath("status", "nodeInfo"),
					k.Status.NodeInfo,
					"nodeInfo is required for in-place upgrade"))
		}
		newSemverVersion, err := version.ParseMajorMinorPatch(v)
		if err != nil {
			allErrs = append(allErrs,
				field.InternalError(
					field.NewPath("metadata", "annotations"),
					errors.Wrapf(err, "failed to parse in-place upgrade version: %s", v)),
			)
		}
		oldSemverVersion, err := version.ParseMajorMinorPatch(k.Status.NodeInfo.KubeletVersion)
		if err != nil {
			allErrs = append(allErrs,
				field.InternalError(
					field.NewPath("status", "nodeInfo", "kubeletVersion"),
					errors.Wrapf(err, "failed to parse old version: %s", k.Status.NodeInfo.KubeletVersion)),
			)
		}

		if newSemverVersion.Equals(oldSemverVersion) {
			allErrs = append(allErrs,
				field.Invalid(
					field.NewPath("metadata", "annotations"),
					v,
					"new version must be different from old version"),
			)
		}

		if !(newSemverVersion.GT(oldSemverVersion) &&
			newSemverVersion.Major == oldSemverVersion.Major &&
			(newSemverVersion.Minor == oldSemverVersion.Minor+1 || newSemverVersion.Minor == oldSemverVersion.Minor)) {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("metadata", "annotations"),
					v, "Skipping MINOR versions when upgrading is unsupported."))
		}
	}

	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *KKInstance) ValidateDelete() (admission.Warnings, error) {
	kkinstancelog.Info("validate delete", "name", k.Name)
	return nil, nil
}
