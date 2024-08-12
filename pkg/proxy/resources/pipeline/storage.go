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

package pipeline

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apiregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
)

type PipelineStorage struct {
	Pipeline       *REST
	PipelineStatus *StatusREST
}

type REST struct {
	*apiregistry.Store
}

type StatusREST struct {
	store *apiregistry.Store
}

func (r *StatusREST) NamespaceScoped() bool {
	return true
}

// New creates a new Node object.
func (r *StatusREST) New() runtime.Object {
	return &kkcorev1.Pipeline{}
}

// Destroy cleans up resources on shutdown.
func (r *StatusREST) Destroy() {
	// Given that underlying store is shared with REST,
	// we don't destroy it here explicitly.
}

// Get retrieves the object from the storage. It is required to support Patch.
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx context.Context, name string, objInfo apirest.UpdatedObjectInfo, createValidation apirest.ValidateObjectFunc, updateValidation apirest.ValidateObjectUpdateFunc, forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {
	// We are explicitly setting forceAllowCreate to false in the call to the underlying storage because
	// subresources should never allow create on update.
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, false, options)
}

// GetResetFields implements rest.ResetFieldsStrategy
func (r *StatusREST) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return r.store.GetResetFields()
}

func (r *StatusREST) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return r.store.ConvertToTable(ctx, object, tableOptions)
}
func NewStorage(optsGetter apigeneric.RESTOptionsGetter) (PipelineStorage, error) {
	store := &apiregistry.Store{
		NewFunc:                   func() runtime.Object { return &kkcorev1.Pipeline{} },
		NewListFunc:               func() runtime.Object { return &kkcorev1.PipelineList{} },
		DefaultQualifiedResource:  kkcorev1.SchemeGroupVersion.WithResource("pipelines").GroupResource(),
		SingularQualifiedResource: kkcorev1.SchemeGroupVersion.WithResource("pipeline").GroupResource(),

		CreateStrategy:      Strategy,
		UpdateStrategy:      Strategy,
		DeleteStrategy:      Strategy,
		ReturnDeletedObject: true,

		TableConvertor: apirest.NewDefaultTableConvertor(kkcorev1.SchemeGroupVersion.WithResource("pipelines").GroupResource()),
	}
	options := &apigeneric.StoreOptions{
		RESTOptions: optsGetter,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return PipelineStorage{}, err
	}

	return PipelineStorage{
		Pipeline:       &REST{store},
		PipelineStatus: &StatusREST{store},
	}, nil
}
