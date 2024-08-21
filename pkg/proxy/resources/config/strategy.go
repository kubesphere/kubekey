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

package config

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apinames "k8s.io/apiserver/pkg/storage/names"
	"sigs.k8s.io/structured-merge-diff/v4/fieldpath"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

// ConfigStrategy implements behavior for Pods
type ConfigStrategy struct {
	runtime.ObjectTyper
	apinames.NameGenerator
}

// Strategy is the default logic that applies when creating and updating Pod
// objects via the REST API.
var Strategy = ConfigStrategy{_const.Scheme, apinames.SimpleNameGenerator}

// ===CreateStrategy===

// NamespaceScoped always true
func (t ConfigStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate do no-thing
func (t ConfigStrategy) PrepareForCreate(context.Context, runtime.Object) {
	// do nothing
}

// Validate always pass
func (t ConfigStrategy) Validate(context.Context, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnCreate do no-thing
func (t ConfigStrategy) WarningsOnCreate(context.Context, runtime.Object) []string {
	// do nothing
	return nil
}

// Canonicalize do no-thing
func (t ConfigStrategy) Canonicalize(runtime.Object) {
	// do nothing
}

// ===UpdateStrategy===

// AllowCreateOnUpdate always false
func (t ConfigStrategy) AllowCreateOnUpdate() bool {
	return false
}

// PrepareForUpdate do no-thing
func (t ConfigStrategy) PrepareForUpdate(context.Context, runtime.Object, runtime.Object) {
	// do nothing
}

// ValidateUpdate do nothing
func (t ConfigStrategy) ValidateUpdate(context.Context, runtime.Object, runtime.Object) field.ErrorList {
	// do nothing
	return nil
}

// WarningsOnUpdate always nil
func (t ConfigStrategy) WarningsOnUpdate(context.Context, runtime.Object, runtime.Object) []string {
	// do nothing
	return nil
}

// AllowUnconditionalUpdate always true
func (t ConfigStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// ===ResetFieldsStrategy===

// GetResetFields always nil
func (t ConfigStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return nil
}
