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
)

// log is for logging in this package.
var kkmachinelog = logf.Log.WithName("kkmachine-resource")

func (k *KKMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachine,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachines,verbs=create;update,versions=v1beta1,name=default.kkmachine.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KKMachine{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (k *KKMachine) Default() {
	kkmachinelog.Info("default", "name", k.Name)

	defaultContainerManager(&k.Spec)
}

func defaultContainerManager(spec *KKMachineSpec) {
	// Direct connection to the user-provided CRI socket
	if spec.ContainerManager.CRISocket != "" {
		return
	}

	if spec.ContainerManager.Type == "" {
		spec.ContainerManager.Type = ContainerdType
	}

	switch spec.ContainerManager.Type {
	case ContainerdType:
		if spec.ContainerManager.Version == "" {
			spec.ContainerManager.Version = DefaultContainerdVersion
		}
		spec.ContainerManager.CRISocket = DefaultContainerdCRISocket
	case DockerType:
		if spec.ContainerManager.Version == "" {
			spec.ContainerManager.Version = DefaultDockerVersion
		}
		spec.ContainerManager.CRISocket = DefaultDockerCRISocket
	}
}

//+kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachines,verbs=create;update,versions=v1beta1,name=validation.kkmachine.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.Validator = &KKMachine{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachine) ValidateCreate() error {
	kkmachinelog.Info("validate create", "name", k.Name)

	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRepository(k.Spec.Repository)...)
	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachine) ValidateUpdate(old runtime.Object) error {
	kkmachinelog.Info("validate update", "name", k.Name)
	var allErrs field.ErrorList
	allErrs = append(allErrs, validateRepository(k.Spec.Repository)...)
	return aggregateObjErrors(k.GroupVersionKind().GroupKind(), k.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (k *KKMachine) ValidateDelete() error {
	kkmachinelog.Info("validate delete", "name", k.Name)
	return nil
}

func validateRepository(repo *Repository) field.ErrorList { //nolint:unparam
	var allErrs field.ErrorList
	if repo == nil {
		return allErrs
	}
	return allErrs
}
