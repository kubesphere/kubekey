package _const

import (
	"context"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// NewTestPlaybook creates a fake controller-runtime client, an Inventory resource with the given hosts,
// and returns the client and a Playbook resource referencing the created Inventory.
// This is intended for use in unit tests.
func NewTestPlaybook(hosts []string) (ctrlclient.Client, *kkcorev1.Playbook, error) {
	// Create a fake client with the required scheme and status subresources.
	client := fake.NewClientBuilder().
		WithScheme(Scheme).
		WithStatusSubresource(&kkcorev1.Playbook{}, &kkcorev1alpha1.Task{}).
		Build()

	// Convert the slice of hostnames to an InventoryHost map.
	inventoryHost := make(kkcorev1.InventoryHost)
	for _, h := range hosts {
		inventoryHost[h] = runtime.RawExtension{}
	}

	// Create an Inventory resource with the generated hosts.
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

	// Persist the Inventory resource using the fake client.
	if err := client.Create(context.TODO(), inventory); err != nil {
		return nil, nil, err
	}

	// Create a Playbook resource that references the created Inventory.
	playbook := &kkcorev1.Playbook{
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
	}

	return client, playbook, nil
}
