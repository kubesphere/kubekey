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

package playbook

import (
	"context"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apiregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	apirest "k8s.io/apiserver/pkg/registry/rest"

	proxy "github.com/kubesphere/kubekey/v4/pkg/proxy/resources"
)

// playbookStorage implements the storage interface for Playbook resource
type playbookStorage struct {
	Playbook       *REST       // Main resource storage
	PlaybookStatus *StatusREST // Status subresource storage
}

// GVK returns the GroupVersionKind of the Playbook resource
func (s *playbookStorage) GVK() schema.GroupVersionKind {
	return kkcorev1.SchemeGroupVersion.WithKind("Playbook")
}

// GVRs returns all GroupVersionResources of the Playbook resource
// Contains main resource and status subresource
func (s *playbookStorage) GVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{
		kkcorev1.SchemeGroupVersion.WithResource("playbooks"),        // Main resource
		kkcorev1.SchemeGroupVersion.WithResource("playbooks/status"), // Status subresource
	}
}

// Storage returns the corresponding REST storage based on GVR
func (s *playbookStorage) Storage(gvr schema.GroupVersionResource) apirest.Storage {
	if gvr.Resource == "playbooks/status" {
		return s.PlaybookStatus
	}
	return s.Playbook
}

// IsAlwaysLocal returns false
// Playbook can be stored in remote Kubernetes cluster or local filesystem
func (s *playbookStorage) IsAlwaysLocal() bool {
	return false
}

// REST is the REST storage wrapper for Playbook main resource
type REST struct {
	*apiregistry.Store
}

// StatusREST is the REST storage implementation for Playbook status subresource
type StatusREST struct {
	store *apiregistry.Store
}

// NamespaceScoped returns true, Playbook is a namespace-scoped resource
func (r *StatusREST) NamespaceScoped() bool {
	return true
}

// New creates a new Playbook object
func (r *StatusREST) New() runtime.Object {
	return &kkcorev1.Playbook{}
}

// Destroy cleans up resources
// Since the underlying store is shared with REST, it is not explicitly destroyed here
func (r *StatusREST) Destroy() {}

// Get retrieves the object (used to support Patch operation)
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update updates the status subresource of an object
func (r *StatusREST) Update(
	ctx context.Context,
	name string,
	objInfo apirest.UpdatedObjectInfo,
	createValidation apirest.ValidateObjectFunc,
	updateValidation apirest.ValidateObjectUpdateFunc,
	_ bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	// Subresources should never allow create on update, forceAllowCreate is set to false
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

// ConvertToTable converts the object to table format
func (r *StatusREST) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return r.store.ConvertToTable(ctx, object, tableOptions)
}

// NewStorage creates the storage for Playbook resource
func NewStorage(optsGetter apigeneric.RESTOptionsGetter) (proxy.ResourceStorage, error) {
	store := &apiregistry.Store{
		NewFunc:                   func() runtime.Object { return &kkcorev1.Playbook{} },
		NewListFunc:               func() runtime.Object { return &kkcorev1.PlaybookList{} },
		DefaultQualifiedResource:  kkcorev1.SchemeGroupVersion.WithResource("playbooks").GroupResource(),
		SingularQualifiedResource: kkcorev1.SchemeGroupVersion.WithResource("playbook").GroupResource(),

		CreateStrategy:      Strategy,
		UpdateStrategy:      Strategy,
		DeleteStrategy:      Strategy,
		ReturnDeletedObject: true,

		TableConvertor: apirest.NewDefaultTableConvertor(kkcorev1.SchemeGroupVersion.WithResource("playbooks").GroupResource()),
	}

	options := &apigeneric.StoreOptions{
		RESTOptions: optsGetter,
	}

	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &playbookStorage{
		Playbook:       &REST{store},
		PlaybookStatus: &StatusREST{store},
	}, nil
}
