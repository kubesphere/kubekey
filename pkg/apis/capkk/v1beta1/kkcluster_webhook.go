package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (k *KKCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(k).
		Complete()
}

// +kubebuilder:webhook:path=/validate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=validation.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1
//

var _ webhook.CustomDefaulter = &KKCluster{}

func (k *KKCluster) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

// +kubebuilder:webhook:path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,name=default.kkcluster.infrastructure.cluster.x-k8s.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &KKCluster{}

func (k *KKCluster) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (k *KKCluster) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (k *KKCluster) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}
