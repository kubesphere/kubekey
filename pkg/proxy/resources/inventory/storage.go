/*
Copyright 2024 The KubeSphere Authors.

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

package inventory

import (
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apiregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	apirest "k8s.io/apiserver/pkg/registry/rest"

	proxy "github.com/kubesphere/kubekey/v4/pkg/proxy/resources"
)

// inventoryStorage implements the storage interface for Inventory resource
type inventoryStorage struct {
	Inventory *REST // Main resource storage
}

// GVK returns the GroupVersionKind of the Inventory resource
func (s *inventoryStorage) GVK() schema.GroupVersionKind {
	return kkcorev1.SchemeGroupVersion.WithKind("Inventory")
}

// GVRs returns all GroupVersionResources of the Inventory resource
func (s *inventoryStorage) GVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{
		kkcorev1.SchemeGroupVersion.WithResource("inventories"),
	}
}

// Storage returns the REST storage for Inventory
func (s *inventoryStorage) Storage(gvr schema.GroupVersionResource) apirest.Storage {
	return s.Inventory
}

// IsAlwaysLocal returns false
// Inventory can be stored in remote Kubernetes cluster or local filesystem
func (s *inventoryStorage) IsAlwaysLocal() bool {
	return false
}

// REST is the REST storage wrapper for Inventory main resource
type REST struct {
	*apiregistry.Store
}

// NewStorage creates the storage for Inventory resource
func NewStorage(optsGetter apigeneric.RESTOptionsGetter) (proxy.ResourceStorage, error) {
	store := &apiregistry.Store{
		NewFunc:                   func() runtime.Object { return &kkcorev1.Inventory{} },
		NewListFunc:               func() runtime.Object { return &kkcorev1.InventoryList{} },
		DefaultQualifiedResource:  kkcorev1.SchemeGroupVersion.WithResource("inventories").GroupResource(),
		SingularQualifiedResource: kkcorev1.SchemeGroupVersion.WithResource("inventory").GroupResource(),

		CreateStrategy:      Strategy,
		UpdateStrategy:      Strategy,
		DeleteStrategy:      Strategy,
		ReturnDeletedObject: true,

		TableConvertor: apirest.NewDefaultTableConvertor(kkcorev1.SchemeGroupVersion.WithResource("inventories").GroupResource()),
	}

	options := &apigeneric.StoreOptions{
		RESTOptions: optsGetter,
	}

	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &inventoryStorage{
		Inventory: &REST{store},
	}, nil
}
