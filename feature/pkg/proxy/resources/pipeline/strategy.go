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
	"errors"
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

// NamespaceScoped always true
func (t pipelineStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate do no-thing
func (t pipelineStrategy) PrepareForCreate(context.Context, runtime.Object) {}

// Validate always pass
func (t pipelineStrategy) Validate(context.Context, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnCreate do no-thing
func (t pipelineStrategy) WarningsOnCreate(context.Context, runtime.Object) []string {
	// do nothing
	return nil
}

// Canonicalize do no-thing
func (t pipelineStrategy) Canonicalize(runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

// AllowCreateOnUpdate always false
func (t pipelineStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate do no-thing
func (t pipelineStrategy) PrepareForUpdate(context.Context, runtime.Object, runtime.Object) {
	// do nothing
}

// ValidateUpdate spec is immutable
func (t pipelineStrategy) ValidateUpdate(_ context.Context, obj, old runtime.Object) field.ErrorList {
	// only support update status
	pipeline, ok := obj.(*kkcorev1.Pipeline)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not Task"))}
	}
	oldPipeline, ok := old.(*kkcorev1.Pipeline)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not Task"))}
	}
	if !reflect.DeepEqual(pipeline.Spec, oldPipeline.Spec) {
		return field.ErrorList{field.Forbidden(field.NewPath("spec"), "spec is immutable")}
	}

	return nil
}

// WarningsOnUpdate always nil
func (t pipelineStrategy) WarningsOnUpdate(context.Context, runtime.Object, runtime.Object) []string {
	// do nothing
	return nil
}

// AllowUnconditionalUpdate always true
func (t pipelineStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

// GetResetFields always nil
func (t pipelineStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}
