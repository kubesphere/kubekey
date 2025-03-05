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
	"errors"
	"reflect"

	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apigeneric "k8s.io/apiserver/pkg/registry/generic"
	apistorage "k8s.io/apiserver/pkg/storage"
	apinames "k8s.io/apiserver/pkg/storage/names"
	cgtoolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const pipelineKind = "Pipeline"

// taskStrategy implements behavior for Pods
type taskStrategy struct {
	runtime.ObjectTyper
	apinames.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod
// objects via the REST API.
var Strategy = taskStrategy{_const.Scheme, apinames.SimpleNameGenerator}

// ===CreateStrategy===

// NamespaceScoped always true
func (t taskStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate set tasks status to pending
func (t taskStrategy) PrepareForCreate(_ context.Context, obj runtime.Object) {
	// init status when create
	if task, ok := obj.(*kkcorev1alpha1.Task); ok {
		task.Status = kkcorev1alpha1.TaskStatus{
			Phase: kkcorev1alpha1.TaskPhasePending,
		}
	}
}

// Validate always pass
func (t taskStrategy) Validate(context.Context, runtime.Object) field.ErrorList {
	return nil
}

// WarningsOnCreate do no-thing
func (t taskStrategy) WarningsOnCreate(context.Context, runtime.Object) []string {
	return nil
}

// Canonicalize do no-thing
func (t taskStrategy) Canonicalize(runtime.Object) {}

// ===UpdateStrategy===

// AllowCreateOnUpdate always false
func (t taskStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate do no-thing
func (t taskStrategy) PrepareForUpdate(context.Context, runtime.Object, runtime.Object) {}

// ValidateUpdate spec is immutable
func (t taskStrategy) ValidateUpdate(_ context.Context, obj, old runtime.Object) field.ErrorList {
	// only support update status
	task, ok := obj.(*kkcorev1alpha1.Task)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not Task"))}
	}
	oldTask, ok := old.(*kkcorev1alpha1.Task)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not Task"))}
	}
	if !reflect.DeepEqual(task.Spec, oldTask.Spec) {
		return field.ErrorList{field.Forbidden(field.NewPath("spec"), "spec is immutable")}
	}

	return nil
}

// WarningsOnUpdate always nil
func (t taskStrategy) WarningsOnUpdate(context.Context, runtime.Object, runtime.Object) []string {
	return nil
}

// AllowUnconditionalUpdate always true
func (t taskStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

// GetResetFields always nil
func (t taskStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}

// OwnerPipelineIndexFunc return value ownerReference.object is pipeline.
func OwnerPipelineIndexFunc(obj any) ([]string, error) {
	task, ok := obj.(*kkcorev1alpha1.Task)
	if !ok {
		return nil, errors.New("not Task")
	}

	var index string
	for _, reference := range task.OwnerReferences {
		if reference.Kind == pipelineKind {
			index = types.NamespacedName{
				Namespace: task.Namespace,
				Name:      reference.Name,
			}.String()

			break
		}
	}
	if index == "" {
		return nil, errors.New("task has no ownerReference.pipeline")
	}

	return []string{index}, nil
}

// Indexers returns the indexers for pod storage.
func Indexers() *cgtoolscache.Indexers {
	return &cgtoolscache.Indexers{
		apistorage.FieldIndex(kkcorev1alpha1.TaskOwnerField): OwnerPipelineIndexFunc,
	}
}

// MatchTask returns a generic matcher for a given label and field selector.
func MatchTask(label labels.Selector, fd fields.Selector) apistorage.SelectionPredicate {
	return apistorage.SelectionPredicate{
		Label:       label,
		Field:       fd,
		GetAttrs:    GetAttrs,
		IndexFields: []string{kkcorev1alpha1.TaskOwnerField},
	}
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	task, ok := obj.(*kkcorev1alpha1.Task)
	if !ok {
		return nil, nil, errors.New("not Task")
	}

	return task.ObjectMeta.Labels, ToSelectableFields(task), nil
}

// ToSelectableFields returns a field set that represents the object
func ToSelectableFields(task *kkcorev1alpha1.Task) fields.Set {
	// The purpose of allocation with a given number of elements is to reduce
	// amount of allocations needed to create the fields.Set. If you add any
	// field here or the number of object-meta related fields changes, this should
	// be adjusted.
	taskSpecificFieldsSet := make(fields.Set)
	for _, reference := range task.OwnerReferences {
		if reference.Kind == pipelineKind {
			taskSpecificFieldsSet[kkcorev1alpha1.TaskOwnerField] = types.NamespacedName{
				Namespace: task.Namespace,
				Name:      reference.Name,
			}.String()

			break
		}
	}

	return apigeneric.AddObjectMetaFieldsSet(taskSpecificFieldsSet, &task.ObjectMeta, true)
}

// OwnerPipelineTriggerFunc returns value ownerReference is pipeline of given object.
func OwnerPipelineTriggerFunc(obj runtime.Object) string {
	if task, ok := obj.(*kkcorev1alpha1.Task); ok {
		for _, reference := range task.OwnerReferences {
			if reference.Kind == pipelineKind {
				return types.NamespacedName{
					Namespace: task.Namespace,
					Name:      reference.Name,
				}.String()
			}
		}
	}

	return ""
}
