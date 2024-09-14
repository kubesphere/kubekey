package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (k *KKMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachines,verbs=create;update,versions=v1beta1,name=validation.kkmachine.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &KKMachine{}

func (k *KKMachine) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkmachine,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkmachines,verbs=create;update,versions=v1beta1,name=default.kkmachine.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &KKMachine{}

func (k *KKMachine) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (k *KKMachine) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (k *KKMachine) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}
