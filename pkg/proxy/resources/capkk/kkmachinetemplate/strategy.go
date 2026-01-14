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

package kkmachinetemplate

import (
	"context"
	"github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	"reflect"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apinames "k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// kkMachineTemplateStrategy implements behavior for Pods
type kkMachineTemplateStrategy struct {
	runtime.ObjectTyper
	apinames.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod
// objects via the REST API.
var Strategy = kkMachineTemplateStrategy{_const.Scheme, apinames.SimpleNameGenerator}

// ===CreateStrategy===

// NamespaceScoped always true
func (t kkMachineTemplateStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate do no-thing
func (t kkMachineTemplateStrategy) PrepareForCreate(context.Context, runtime.Object) {}

// Validate always pass
func (t kkMachineTemplateStrategy) Validate(context.Context, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnCreate do no-thing
func (t kkMachineTemplateStrategy) WarningsOnCreate(context.Context, runtime.Object) []string {
	// do nothing
	return nil
}

// Canonicalize do no-thing
func (t kkMachineTemplateStrategy) Canonicalize(runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

// AllowCreateOnUpdate always false
func (t kkMachineTemplateStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate do no-thing
func (t kkMachineTemplateStrategy) PrepareForUpdate(context.Context, runtime.Object, runtime.Object) {
	// do nothing
}

// ValidateUpdate spec is immutable
func (t kkMachineTemplateStrategy) ValidateUpdate(_ context.Context, obj, old runtime.Object) field.ErrorList {
	// only support update status
	item, ok := obj.(*v1beta1.KKMachineTemplate)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not KKMachineTemplate"))}
	}
	oldItem, ok := old.(*v1beta1.KKMachineTemplate)
	if !ok {
		return field.ErrorList{field.InternalError(field.NewPath("spec"), errors.New("the object is not KKMachineTemplate"))}
	}
	if !reflect.DeepEqual(item.Spec, oldItem.Spec) {
		return field.ErrorList{field.Forbidden(field.NewPath("spec"), "spec is immutable")}
	}

	return nil
}

// WarningsOnUpdate always nil
func (t kkMachineTemplateStrategy) WarningsOnUpdate(context.Context, runtime.Object, runtime.Object) []string {
	// do nothing
	return nil
}

// AllowUnconditionalUpdate always true
func (t kkMachineTemplateStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

// GetResetFields always nil
func (t kkMachineTemplateStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}
