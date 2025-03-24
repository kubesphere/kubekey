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

package v1alpha1

import (
	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// TaskOwnerField is the field name of the owner reference in the task.
// It defined in proxy transport. Not applicable in kube-apiserver.
const TaskOwnerField = "ownerReferences:playbook"

// AddConversionFuncs adds the conversion functions to the given scheme.
// NOTE: ownerReferences:playbook is valid in proxy client.
func AddConversionFuncs(scheme *runtime.Scheme) error {
	return scheme.AddFieldLabelConversionFunc(
		SchemeGroupVersion.WithKind("Task"),
		func(label, value string) (string, string, error) {
			switch label {
			case "metadata.name", "metadata.namespace", TaskOwnerField:
				return label, value, nil
			default:
				return "", "", errors.Errorf("field label %q not supported for Task", label)
			}
		},
	)
}
