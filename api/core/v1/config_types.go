/*
Copyright 2023 The KubeSphere Authors.

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

package v1

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Config store global vars for playbook.
type Config struct {
	metav1.TypeMeta `json:",inline"`

	Spec runtime.RawExtension `json:"spec,omitempty"`
}

// UnmarshalJSON decodes spec.Raw into spec.Object
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	aux := &Alias{}
	if err := json.Unmarshal(data, aux); err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}
	*c = Config(*aux)

	// Decode spec.Raw into spec.Object if it's not already set
	objMap := make(map[string]any)
	if len(c.Spec.Raw) > 0 && c.Spec.Object == nil {
		if err := json.Unmarshal(c.Spec.Raw, &objMap); err != nil {
			return errors.Wrap(err, "failed to unmarshal spec.Raw")
		}
	}
	c.Spec.Object = &unstructured.Unstructured{Object: objMap}

	return nil
}

// MarshalJSON ensures spec.Object is converted back to spec.Raw
func (c *Config) MarshalJSON() ([]byte, error) {
	// Ensure spec.Object is serialized into spec.Raw
	if c.Spec.Object != nil {
		raw, err := json.Marshal(c.Spec.Object)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal spec.Object")
		}
		c.Spec.Raw = raw
	}

	type Alias Config
	return json.Marshal((*Alias)(c))
}

// Value returns the underlying map[string]any from the Config's unstructured Object.
// This provides direct access to the config values stored in Spec.Object.
func (c *Config) Value() map[string]any {
	if c.Spec.Object == nil {
		return make(map[string]any)
	}

	return c.Spec.Object.(*unstructured.Unstructured).Object
}
