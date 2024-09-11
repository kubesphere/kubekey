package executor

import (
	"context"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

func newTestOption() (*option, error) {
	var err error

	o := &option{
		client: fake.NewClientBuilder().WithScheme(_const.Scheme).WithStatusSubresource(&kkcorev1.Pipeline{}, &kkcorev1alpha1.Task{}).Build(),
		pipeline: &kkcorev1.Pipeline{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: corev1.NamespaceDefault,
			},
			Spec: kkcorev1.PipelineSpec{
				InventoryRef: &corev1.ObjectReference{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
				},
				ConfigRef: &corev1.ObjectReference{
					Name:      "test",
					Namespace: corev1.NamespaceDefault,
				},
			},
			Status: kkcorev1.PipelineStatus{},
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

	if err := o.client.Create(context.TODO(), &kkcorev1.Config{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: corev1.NamespaceDefault,
		},
		Spec: runtime.RawExtension{},
	}); err != nil {
		return nil, err
	}

	o.variable, err = variable.New(context.TODO(), o.client, *o.pipeline, source.MemorySource)
	if err != nil {
		return nil, err
	}

	return o, nil
}
