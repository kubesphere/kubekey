package executor

import (
	"context"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

func newTestOption(hosts []string) (*option, error) {
	var err error
	// convert host to InventoryHost
	inventoryHost := make(kkcorev1.InventoryHost)
	for _, h := range hosts {
		inventoryHost[h] = runtime.RawExtension{}
	}
	client := fake.NewClientBuilder().WithScheme(_const.Scheme).WithStatusSubresource(&kkcorev1.Playbook{}, &kkcorev1alpha1.Task{}).Build()
	inventory := &kkcorev1.Inventory{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-",
			Namespace:    corev1.NamespaceDefault,
		},
		Spec: kkcorev1.InventorySpec{
			Hosts: inventoryHost,
		},
	}
	if err := client.Create(context.TODO(), inventory); err != nil {
		return nil, err
	}

	o := &option{
		client: client,
		playbook: &kkcorev1.Playbook{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-",
				Namespace:    corev1.NamespaceDefault,
			},
			Spec: kkcorev1.PlaybookSpec{
				InventoryRef: &corev1.ObjectReference{
					Name:      inventory.Name,
					Namespace: inventory.Namespace,
				},
			},
			Status: kkcorev1.PlaybookStatus{},
		},
		logOutput: os.Stdout,
	}

	o.variable, err = variable.New(context.TODO(), o.client, *o.playbook, source.MemorySource)
	if err != nil {
		return nil, err
	}

	return o, nil
}
