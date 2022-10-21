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
)

var (
	conflictingFileSourceMsg = "only one of content or contentFrom may be specified for a single file"
	missingSecretNameMsg     = "secret file source must specify non-empty secret name"
	missingSecretKeyMsg      = "secret file source must specify non-empty secret key"
	pathConflictMsg          = "path property must be unique among all files"
)

func (c *K3sConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(c).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/mutate-bootstrap-cluster-x-k8s-io-v1beta1-k3sconfig,mutating=true,failurePolicy=fail,sideEffects=None,groups=bootstrap.cluster.x-k8s.io,resources=k3sconfigs,versions=v1beta1,name=default.k3sconfig.bootstrap.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Defaulter = &K3sConfig{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (c *K3sConfig) Default() {
	DefaultK3sConfigSpec(&c.Spec)
}

// DefaultK3sConfigSpec defaults a K3sConfigSpec.
func DefaultK3sConfigSpec(c *K3sConfigSpec) {
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-bootstrap-cluster-x-k8s-io-v1beta1-k3sconfig,mutating=false,failurePolicy=fail,matchPolicy=Equivalent,groups=bootstrap.cluster.x-k8s.io,resources=k3sconfigs,versions=v1beta1,name=validation.k3sconfig.bootstrap.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

var _ webhook.Validator = &K3sConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (c *K3sConfig) ValidateCreate() error {
	return c.Spec.validate(c.Name)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (c *K3sConfig) ValidateUpdate(old runtime.Object) error {
	return c.Spec.validate(c.Name)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (c *K3sConfig) ValidateDelete() error {
	return nil
}

func (c *K3sConfigSpec) validate(name string) error {
	allErrs := c.Validate(field.NewPath("spec"))

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind("K3sConfig").GroupKind(), name, allErrs)
}

// Validate ensures the K3sConfigSpec is valid.
func (c *K3sConfigSpec) Validate(pathPrefix *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, c.validateFiles(pathPrefix)...)

	return allErrs
}

func (c *K3sConfigSpec) validateFiles(pathPrefix *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	knownPaths := map[string]struct{}{}

	for i := range c.Files {
		file := c.Files[i]
		if file.Content != "" && file.ContentFrom != nil {
			allErrs = append(
				allErrs,
				field.Invalid(
					pathPrefix.Child("files").Index(i),
					file,
					conflictingFileSourceMsg,
				),
			)
		}
		// n.b.: if we ever add types besides Secret as a ContentFrom
		// Source, we must add webhook validation here for one of the
		// sources being non-nil.
		if file.ContentFrom != nil {
			if file.ContentFrom.Secret.Name == "" {
				allErrs = append(
					allErrs,
					field.Required(
						pathPrefix.Child("files").Index(i).Child("contentFrom", "secret", "name"),
						missingSecretNameMsg,
					),
				)
			}
			if file.ContentFrom.Secret.Key == "" {
				allErrs = append(
					allErrs,
					field.Required(
						pathPrefix.Child("files").Index(i).Child("contentFrom", "secret", "key"),
						missingSecretKeyMsg,
					),
				)
			}
		}
		_, conflict := knownPaths[file.Path]
		if conflict {
			allErrs = append(
				allErrs,
				field.Invalid(
					pathPrefix.Child("files").Index(i).Child("path"),
					file,
					pathConflictMsg,
				),
			)
		}
		knownPaths[file.Path] = struct{}{}
	}

	return allErrs
}
