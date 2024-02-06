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

package proxy

import (
	"context"

	"k8s.io/apiserver/pkg/admission"
)

func newAlwaysAdmit() admission.Interface {
	return &admit{}
}

type admit struct {
}

func (a admit) Validate(ctx context.Context, attr admission.Attributes, obj admission.ObjectInterfaces) (err error) {
	return nil
}

func (a admit) Admit(ctx context.Context, attr admission.Attributes, obj admission.ObjectInterfaces) (err error) {
	return nil
}

func (a admit) Handles(operation admission.Operation) bool {
	return true
}

var _ admission.MutationInterface = admit{}
var _ admission.ValidationInterface = admit{}
