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

package task

import (
	"context"

	"github.com/cockroachdb/errors"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apiregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	apistorage "k8s.io/apiserver/pkg/storage"

	proxy "github.com/kubesphere/kubekey/v4/pkg/proxy/resources"
)

// taskStorage implements the storage interface for Task resource
type taskStorage struct {
	Task       *REST       // Main resource storage
	TaskStatus *StatusREST // Status subresource storage
}

// GVK returns the GroupVersionKind of the Task resource
func (s *taskStorage) GVK() schema.GroupVersionKind {
	return kkcorev1alpha1.SchemeGroupVersion.WithKind("Task")
}

// GVRs returns all GroupVersionResources of the Task resource
// Contains main resource and status subresource
func (s *taskStorage) GVRs() []schema.GroupVersionResource {
	return []schema.GroupVersionResource{
		kkcorev1alpha1.SchemeGroupVersion.WithResource("tasks"),        // Main resource
		kkcorev1alpha1.SchemeGroupVersion.WithResource("tasks/status"), // Status subresource
	}
}

// Storage returns the corresponding REST storage based on GVR
func (s *taskStorage) Storage(gvr schema.GroupVersionResource) apirest.Storage {
	if gvr.Resource == "tasks/status" {
		return s.TaskStatus
	}
	return s.Task
}

// IsAlwaysLocal returns true
// Task is running data, large volume and requires reentrancy, always uses local file storage
func (s *taskStorage) IsAlwaysLocal() bool {
	return true
}

// REST is the REST storage wrapper for Task main resource
type REST struct {
	*apiregistry.Store
}

// StatusREST is the REST storage implementation for Task status subresource
type StatusREST struct {
	store *apiregistry.Store
}

// NamespaceScoped returns true, Task is a namespace-scoped resource
func (r *StatusREST) NamespaceScoped() bool {
	return true
}

// New creates a new Task object
func (r *StatusREST) New() runtime.Object {
	return &kkcorev1alpha1.Task{}
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

// NewStorage creates the storage for Task resource
func NewStorage(optsGetter apigeneric.RESTOptionsGetter) (proxy.ResourceStorage, error) {
	store := &apiregistry.Store{
		NewFunc:                   func() runtime.Object { return &kkcorev1alpha1.Task{} },
		NewListFunc:               func() runtime.Object { return &kkcorev1alpha1.TaskList{} },
		PredicateFunc:             MatchTask,
		DefaultQualifiedResource:  kkcorev1alpha1.SchemeGroupVersion.WithResource("tasks").GroupResource(),
		SingularQualifiedResource: kkcorev1alpha1.SchemeGroupVersion.WithResource("task").GroupResource(),

		CreateStrategy:      Strategy,
		UpdateStrategy:      Strategy,
		DeleteStrategy:      Strategy,
		ReturnDeletedObject: true,

		TableConvertor: apirest.NewDefaultTableConvertor(kkcorev1alpha1.SchemeGroupVersion.WithResource("tasks").GroupResource()),
	}

	options := &apigeneric.StoreOptions{
		RESTOptions: optsGetter,
		AttrFunc:    GetAttrs,
		TriggerFunc: map[string]apistorage.IndexerFunc{"metadata.name": NameTriggerFunc},
	}

	if err := store.CompleteWithOptions(options); err != nil {
		return nil, errors.Wrap(err, "failed to complete store")
	}

	return &taskStorage{
		Task:       &REST{store},
		TaskStatus: &StatusREST{store},
	}, nil
}
