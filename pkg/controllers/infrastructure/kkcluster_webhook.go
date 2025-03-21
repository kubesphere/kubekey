package infrastructure

import (
	"context"

	"github.com/cockroachdb/errors"
	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// +kubebuilder:webhookconfiguration:mutating=true,name=default-capkk
// +kubebuilder:webhook:mutating=true,name=default.kkcluster.infrastructure.cluster.x-k8s.io,serviceName=capkk-webhook-service,serviceNamespace=capkk-system,path=/mutate-infrastructure-cluster-x-k8s-io-v1beta1-kkcluster,failurePolicy=fail,sideEffects=None,groups=infrastructure.cluster.x-k8s.io,resources=kkclusters,verbs=create;update,versions=v1beta1,admissionReviewVersions=v1

// KKClusterWebhook reconciles a KKCluster object
type KKClusterWebhook struct {
}

var _ admission.CustomDefaulter = &KKClusterWebhook{}
var _ options.Controller = &KKClusterWebhook{}

// Default implements admission.CustomDefaulter.
func (w *KKClusterWebhook) Default(ctx context.Context, obj runtime.Object) error {
	kkcluster, ok := obj.(*capkkinfrav1beta1.KKCluster)
	if !ok {
		return errors.New("cannot convert to kkclusters")
	}
	if kkcluster.Spec.HostCheckGroup == "" {
		kkcluster.Spec.HostCheckGroup = _const.VariableUnGrouped
	}
	if kkcluster.Spec.ControlPlaneEndpointType == "" {
		kkcluster.Spec.ControlPlaneEndpointType = capkkinfrav1beta1.ControlPlaneEndpointTypeVIP
	}

	return nil
}

// Name implements controllers.Controller.
func (w *KKClusterWebhook) Name() string {
	return "kkcluster-webhook"
}

func (w *KKClusterWebhook) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	return ctrl.NewWebhookManagedBy(mgr).
		WithDefaulter(w).
		For(&capkkinfrav1beta1.KKCluster{}).
		Complete()
}
