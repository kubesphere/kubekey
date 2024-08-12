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
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apinames "k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// pipelineStrategy implements behavior for Pods
type pipelineStrategy struct {
	runtime.ObjectTyper
	apinames.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod
// objects via the REST API.
var Strategy = pipelineStrategy{_const.Scheme, apinames.SimpleNameGenerator}

// ===CreateStrategy===

func (t pipelineStrategy) NamespaceScoped() bool {
	return true
}

func (t pipelineStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	// do nothing
}

func (t pipelineStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

func (t pipelineStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	// do nothing
	return nil
}

func (t pipelineStrategy) Canonicalize(obj runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

func (t pipelineStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (t pipelineStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	// do nothing
}

func (t pipelineStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	// only support update status
	task := obj.(*kkcorev1.Pipeline)
	oldTask := old.(*kkcorev1.Pipeline)
	if !reflect.DeepEqual(task.Spec, oldTask.Spec) {
		return field.ErrorList{field.Forbidden(field.NewPath("spec"), "spec is immutable")}
	}
	return nil
}

func (t pipelineStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	// do nothing
	return nil
}

func (t pipelineStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

func (t pipelineStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}
