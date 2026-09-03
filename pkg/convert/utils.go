/*
Copyright 2026 The KubeSphere Authors.

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

package convert

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/runtime"
	sigsyaml "sigs.k8s.io/yaml"
)

// toRawExtension serializes a vars map into a runtime.RawExtension.
func toRawExtension(vars map[string]any) (runtime.RawExtension, error) {
	raw, err := json.Marshal(vars)
	if err != nil {
		return runtime.RawExtension{}, errors.Wrap(err, "failed to marshal host vars")
	}
	return runtime.RawExtension{Raw: raw}, nil
}

// marshalYAML marshals an object to YAML via JSON so that json tags apply.
func marshalYAML(obj any) ([]byte, error) {
	return sigsyaml.Marshal(obj)
}
