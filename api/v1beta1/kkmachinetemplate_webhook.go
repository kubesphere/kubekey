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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var kkmachinetemplatelog = logf.Log.WithName("kkmachinetemplate-resource")

func (k *KKMachineTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachinetemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachinetemplates,verbs=create;update,versions=v1beta1,name=default.kkmachinetemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKMachineTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *KKMachineTemplate) Default() {
	kkmachinetemplatelog.Info("default", "name", k.Name)

	defaultContainerManager(&k.Spec.Template.Spec)
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachinetemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachinetemplates,verbs=create;update,versions=v1beta1,name=validation.kkmachinetemplate.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKMachineTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachineTemplate) ValidateCreate() (admission.Warnings, error) {
	kkmachinetemplatelog.Info("validate create", "name", k.Name)

	spec := k.Spec.Template.Spec
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRepository(spec.Repository)...)
	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachineTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	kkmachinetemplatelog.Info("validate update", "name", k.Name)

	spec := k.Spec.Template.Spec
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRepository(spec.Repository)...)
	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachineTemplate) ValidateDelete() (admission.Warnings, error) {
	kkmachinetemplatelog.Info("validate delete", "name", k.Name)

	return nil, nil
}
