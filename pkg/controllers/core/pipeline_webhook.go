package core

import (
	"context"
	"errors"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/kubesphere/kubekey/v4/cmd/controller-manager/app/options"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const (
	// defaultServiceAccountName is the default serviceaccount name for pipeline's executor pod.
	defaultServiceAccountName = "kubekey-executor"
	// defaultServiceAccountName is the default clusterrolebinding name for defaultServiceAccountName.
	defaultClusterRoleBindingName = "kubekey-executor"
)

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:webhook:mutating=true,name=default.pipeline.kubekey.kubesphere.io,serviceName=kk-webhook-service,serviceNamespace=capkk-system,path=/mutate-kubekey-kubesphere-io-v1-pipeline,failurePolicy=fail,sideEffects=None,groups=kubekey.kubesphere.io,resources=pipelines,verbs=create;update,versions=v1,admissionReviewVersions=v1

// PipelineWebhook handles mutating webhooks for Pipelines.
type PipelineWebhook struct {
	ctrlclient.Client
}

var _ admission.CustomDefaulter = &PipelineWebhook{}
var _ options.Controller = &PipelineWebhook{}

// Name implements controllers.Controller.
func (w *PipelineWebhook) Name() string {
	return "pipeline-webhook"
}

// SetupWithManager implements controllers.Controller.
func (w *PipelineWebhook) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	w.Client = mgr.GetClient()

	return ctrl.NewWebhookManagedBy(mgr).
		WithDefaulter(w).
		For(&kkcorev1.Pipeline{}).
		Complete()
}

// Default implements admission.CustomDefaulter.
func (w *PipelineWebhook) Default(ctx context.Context, obj runtime.Object) error {
	pipeline, ok := obj.(*kkcorev1.Pipeline)
	if !ok {
		return errors.New("cannot convert to pipelines")
	}
	if pipeline.Spec.ServiceAccountName == "" && os.Getenv(_const.ENV_EXECUTOR_CLUSTERROLE) != "" {
		// should create default service account in current namespace
		if err := w.syncServiceAccount(ctx, pipeline, os.Getenv(_const.ENV_EXECUTOR_CLUSTERROLE)); err != nil {
			return err
		}
		pipeline.Spec.ServiceAccountName = defaultServiceAccountName
	}
	if pipeline.Spec.ServiceAccountName == "" {
		pipeline.Spec.ServiceAccountName = "default"
	}

	return nil
}

func (w *PipelineWebhook) syncServiceAccount(ctx context.Context, pipeline *kkcorev1.Pipeline, clusterrole string) error {
	// check if clusterrole is exist
	cr := &rbacv1.ClusterRole{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Name: clusterrole}, cr); err != nil {
		return err
	}

	// check if the default service account is exist
	sa := &corev1.ServiceAccount{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: pipeline.Namespace, Name: defaultServiceAccountName}, sa); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: pipeline.Namespace,
				Name:      defaultServiceAccountName,
			},
		}
		if err := w.Client.Create(ctx, sa); err != nil {
			return err
		}
	}
	// check if the service account is bound to the default cluster role
	crb := &rbacv1.ClusterRoleBinding{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Name: defaultClusterRoleBindingName}, crb); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}

		// create clusterrolebinding
		return w.Client.Create(ctx, &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultClusterRoleBindingName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      defaultServiceAccountName,
					Namespace: pipeline.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: clusterrole,
			},
		})
	}

	for _, sj := range crb.Subjects {
		if sj.Kind == "ServiceAccount" && sj.Name == defaultServiceAccountName && sj.Namespace == pipeline.Namespace {
			return nil
		}
	}
	ncrb := crb.DeepCopy()
	ncrb.Subjects = append(crb.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      defaultServiceAccountName,
		Namespace: pipeline.Namespace,
	})

	return w.Client.Patch(ctx, ncrb, ctrlclient.MergeFrom(crb))
}
