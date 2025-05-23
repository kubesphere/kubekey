package core

import (
	"context"

	"github.com/cockroachdb/errors"
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
	// defaultServiceAccountName is the default serviceaccount name for playbook's executor pod.
	defaultServiceAccountName = "kubekey-executor"
	// defaultServiceAccountName is the default clusterrolebinding name for defaultServiceAccountName.
	defaultClusterRoleBindingName = "kubekey-executor"
)

// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:webhook:mutating=true,name=default.playbook.kubekey.kubesphere.io,serviceName=kk-webhook-service,serviceNamespace=capkk-system,path=/mutate-kubekey-kubesphere-io-v1-playbook,failurePolicy=fail,sideEffects=None,groups=kubekey.kubesphere.io,resources=playbooks,verbs=create;update,versions=v1,admissionReviewVersions=v1

// PlaybookWebhook handles mutating webhooks for Playbooks.
type PlaybookWebhook struct {
	ctrlclient.Client
}

var _ admission.CustomDefaulter = &PlaybookWebhook{}
var _ options.Controller = &PlaybookWebhook{}

// Name implements controllers.Controller.
func (w *PlaybookWebhook) Name() string {
	return "playbook-webhook"
}

// SetupWithManager implements controllers.Controller.
func (w *PlaybookWebhook) SetupWithManager(mgr ctrl.Manager, o options.ControllerManagerServerOptions) error {
	w.Client = mgr.GetClient()

	return ctrl.NewWebhookManagedBy(mgr).
		WithDefaulter(w).
		For(&kkcorev1.Playbook{}).
		Complete()
}

// Default implements admission.CustomDefaulter.
func (w *PlaybookWebhook) Default(ctx context.Context, obj runtime.Object) error {
	playbook, ok := obj.(*kkcorev1.Playbook)
	if !ok {
		return errors.Errorf("failed to convert %q to playbooks", obj.GetObjectKind().GroupVersionKind().String())
	}
	if playbook.Spec.ServiceAccountName == "" && _const.Getenv(_const.ExecutorClusterRole) != "" {
		// should create default service account in current namespace
		if err := w.syncServiceAccount(ctx, playbook, _const.Getenv(_const.ExecutorClusterRole)); err != nil {
			return err
		}
		playbook.Spec.ServiceAccountName = defaultServiceAccountName
	}
	if playbook.Spec.ServiceAccountName == "" {
		playbook.Spec.ServiceAccountName = "default"
	}

	return nil
}

func (w *PlaybookWebhook) syncServiceAccount(ctx context.Context, playbook *kkcorev1.Playbook, clusterrole string) error {
	// check if clusterrole is exist
	cr := &rbacv1.ClusterRole{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Name: clusterrole}, cr); err != nil {
		return errors.Wrapf(err, "failed to get clusterrole %q for playbook %q", clusterrole, ctrlclient.ObjectKeyFromObject(playbook))
	}
	// check if the default service account is exist
	sa := &corev1.ServiceAccount{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Namespace: playbook.Namespace, Name: defaultServiceAccountName}, sa); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to get serviceaccount %q for playbook %q", defaultServiceAccountName, ctrlclient.ObjectKeyFromObject(playbook))
		}
		// create service account if not exist.
		sa = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: playbook.Namespace,
				Name:      defaultServiceAccountName,
			},
		}
		if err := w.Client.Create(ctx, sa); err != nil {
			return errors.WithStack(err)
		}
	}
	// check if the service account is bound to the default cluster role
	crb := &rbacv1.ClusterRoleBinding{}
	if err := w.Client.Get(ctx, ctrlclient.ObjectKey{Name: defaultClusterRoleBindingName}, crb); err != nil {
		if !apierrors.IsNotFound(err) {
			return errors.Wrapf(err, "failed to get clusterrolebinding %q for playbook %q", defaultClusterRoleBindingName, ctrlclient.ObjectKeyFromObject(playbook))
		}
		// create clusterrolebinding if not exist
		return w.Client.Create(ctx, &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultClusterRoleBindingName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      defaultServiceAccountName,
					Namespace: playbook.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: clusterrole,
			},
		})
	}

	for _, sj := range crb.Subjects {
		if sj.Kind == "ServiceAccount" && sj.Name == defaultServiceAccountName && sj.Namespace == playbook.Namespace {
			return nil
		}
	}
	ncrb := crb.DeepCopy()
	ncrb.Subjects = append(crb.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      defaultServiceAccountName,
		Namespace: playbook.Namespace,
	})

	return errors.Wrapf(w.Client.Patch(ctx, ncrb, ctrlclient.MergeFrom(crb)),
		"fail to update clusterrolebinding %q for playbook %q", defaultClusterRoleBindingName, ctrlclient.ObjectKeyFromObject(playbook))
}
