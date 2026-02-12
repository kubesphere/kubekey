/*
Copyright 2023 The KubeSphere Authors.

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

package _const

import (
	"context"
	"fmt"
	"os"
	"strings"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// GetWorkdirFromConfig retrieves the working directory from the provided configuration.
// If the 'workdir' value is set in the configuration and is a string, it returns that value.
// If the 'workdir' value is not set or is not a string, it logs an informational message
// and attempts to get the current working directory of the process.
// If it fails to get the current working directory, it logs another informational message
// and returns a default directory path "/opt/kubekey".
func GetWorkdirFromConfig(config kkcorev1.Config) string {
	if wd, _, err := unstructured.NestedString(config.Value(), Workdir); err == nil {
		return wd
	}
	klog.V(2).Info("work_dir is not set, using current directory")
	wd, err := os.Getwd()
	if err != nil {
		klog.V(2).Info("failed to get current dir, using default", "default", "/opt/kubekey")
		return "/opt/kubekey"
	}

	return wd
}

// Host2ProviderID converts a cluster name and host into a provider ID string.
// It returns a pointer to a string in the format "kk://<cluster_name>/<host>".
func Host2ProviderID(clusterName, host string) *string {
	return ptr.To(fmt.Sprintf("kk://%s/%s", clusterName, host))
}

// ProviderID2Host extracts the host name from a provider ID string.
// It takes a cluster name and provider ID pointer, and returns the host portion
// by trimming off the "kk://<cluster_name>/" prefix. If providerID is nil,
// returns an empty string.
func ProviderID2Host(clusterName string, providerID *string) string {
	return strings.TrimPrefix(ptr.Deref(providerID, ""), fmt.Sprintf("kk://%s/", clusterName))
}

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
