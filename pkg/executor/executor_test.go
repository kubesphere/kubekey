package executor

import (
	"context"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

func newTestOption() (*option, error) {
	var err error

	o := &option{
		client: fake.NewClientBuilder().WithScheme(_const.Scheme).WithStatusSubresource(&kkcorev1.Playbook{}, &kkcorev1alpha1.Task{}).Build(),
		playbook: &kkcorev1.Playbook{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: corev1.NamespaceDefault,
			},
			Spec: kkcorev1.PlaybookSpec{
				InventoryRef: &corev1.ObjectReference{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
				},
			},
			Status: kkcorev1.PlaybookStatus{},
		},
		logOutput: os.Stdout,
	}

	if err := o.client.Create(context.TODO(), &kkcorev1.Inventory{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: corev1.NamespaceDefault,
		},
		Spec: kkcorev1.InventorySpec{},
	}); err != nil {
		return nil, err
	}

	o.variable, err = variable.New(context.TODO(), o.client, *o.playbook, source.MemorySource)
	if err != nil {
		return nil, err
	}

	return o, nil
}
