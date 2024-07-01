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
	"fmt"
	"reflect"

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

	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
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

func (t taskStrategy) NamespaceScoped() bool {
	return true
}

func (t taskStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	// init status when create
	task := obj.(*kubekeyv1alpha1.Task)
	task.Status = kubekeyv1alpha1.TaskStatus{
		Phase: kubekeyv1alpha1.TaskPhasePending,
	}
}

func (t taskStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

func (t taskStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	// do nothing
	return nil
}

func (t taskStrategy) Canonicalize(obj runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

func (t taskStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (t taskStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	// do nothing
}

func (t taskStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	// only support update status
	task := obj.(*kubekeyv1alpha1.Task)
	oldTask := old.(*kubekeyv1alpha1.Task)
	if !reflect.DeepEqual(task.Spec, oldTask.Spec) {
		return field.ErrorList{field.Forbidden(field.NewPath("spec"), "spec is immutable")}
	}
	return nil
}

func (t taskStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	// do nothing
	return nil
}

func (t taskStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

func (t taskStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}

// OwnerPipelineIndexFunc return value ownerReference.object is pipeline.
func OwnerPipelineIndexFunc(obj interface{}) ([]string, error) {
	task, ok := obj.(*kubekeyv1alpha1.Task)
	if !ok {
		return nil, fmt.Errorf("not a task")
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
		return nil, fmt.Errorf("task has no ownerReference.pipeline")
	}

	return []string{index}, nil
}

// Indexers returns the indexers for pod storage.
func Indexers() *cgtoolscache.Indexers {
	return &cgtoolscache.Indexers{
		apistorage.FieldIndex(kubekeyv1alpha1.TaskOwnerField): OwnerPipelineIndexFunc,
	}
}

// MatchTask returns a generic matcher for a given label and field selector.
func MatchTask(label labels.Selector, field fields.Selector) apistorage.SelectionPredicate {
	return apistorage.SelectionPredicate{
		Label:       label,
		Field:       field,
		GetAttrs:    GetAttrs,
		IndexFields: []string{kubekeyv1alpha1.TaskOwnerField},
	}
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	task, ok := obj.(*kubekeyv1alpha1.Task)
	if !ok {
		return nil, nil, fmt.Errorf("not a Task")
	}
	return labels.Set(task.ObjectMeta.Labels), ToSelectableFields(task), nil
}

// ToSelectableFields returns a field set that represents the object
// TODO: fields are not labels, and the validation rules for them do not apply.
func ToSelectableFields(task *kubekeyv1alpha1.Task) fields.Set {
	// The purpose of allocation with a given number of elements is to reduce
	// amount of allocations needed to create the fields.Set. If you add any
	// field here or the number of object-meta related fields changes, this should
	// be adjusted.
	taskSpecificFieldsSet := make(fields.Set, 10)
	for _, reference := range task.OwnerReferences {
		if reference.Kind == pipelineKind {
			taskSpecificFieldsSet[kubekeyv1alpha1.TaskOwnerField] = types.NamespacedName{
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
	task := obj.(*kubekeyv1alpha1.Task)
	for _, reference := range task.OwnerReferences {
		if reference.Kind == pipelineKind {
			return types.NamespacedName{
				Namespace: task.Namespace,
				Name:      reference.Name,
			}.String()
		}
	}
	return ""
}
