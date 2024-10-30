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
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apinames "k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// inventoryStrategy implements behavior for Pods
type inventoryStrategy struct {
	runtime.ObjectTyper
	apinames.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod
// objects via the REST API.
var Strategy = inventoryStrategy{_const.Scheme, apinames.SimpleNameGenerator}

// ===CreateStrategy===

// NamespaceScoped always true
func (t inventoryStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate do no-thing
func (t inventoryStrategy) PrepareForCreate(context.Context, runtime.Object) {
	// do nothing
}

// Validate always pass
func (t inventoryStrategy) Validate(context.Context, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnCreate do no-thing
func (t inventoryStrategy) WarningsOnCreate(context.Context, runtime.Object) []string {
	// do nothing
	return nil
}

// Canonicalize do no-thing
func (t inventoryStrategy) Canonicalize(runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

// AllowCreateOnUpdate always false
func (t inventoryStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate do no-thing
func (t inventoryStrategy) PrepareForUpdate(context.Context, runtime.Object, runtime.Object) {}

// ValidateUpdate do nothing
func (t inventoryStrategy) ValidateUpdate(context.Context, runtime.Object, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnUpdate always nil
func (t inventoryStrategy) WarningsOnUpdate(context.Context, runtime.Object, runtime.Object) []string {
	// do nothing
	return nil
}

// AllowUnconditionalUpdate always true
func (t inventoryStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

// GetResetFields always nil
func (t inventoryStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}
