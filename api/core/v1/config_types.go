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
	"reflect"
	"strings"

	"github.com/cockroachdb/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

// Config store global vars for playbook.
type Config struct {
	metav1.TypeMeta `json:",inline"`

	Spec runtime.RawExtension `json:"spec,omitempty"`
}

// SetValue to config
// if key contains "." (a.b), will convert map and set value (a:b:value)
func (c *Config) SetValue(key string, value any) error {
	configMap := make(map[string]any)
	if c.Spec.Raw != nil {
		if err := json.Unmarshal(c.Spec.Raw, &configMap); err != nil {
			return errors.WithStack(err)
		}
	}
	// set value
	var f func(input map[string]any, key []string, value any) any
	f = func(input map[string]any, key []string, value any) any {
		if len(key) == 0 {
			return input
		}

		firstKey := key[0]
		if len(key) == 1 {
			input[firstKey] = value

			return input
		}

		// Handle nested maps
		if v, ok := input[firstKey]; ok && reflect.TypeOf(v).Kind() == reflect.Map {
			if vd, ok := v.(map[string]any); ok {
				input[firstKey] = f(vd, key[1:], value)
			}
		} else {
			input[firstKey] = f(make(map[string]any), key[1:], value)
		}

		return input
	}
	data, err := json.Marshal(f(configMap, strings.Split(key, "."), value))
	if err != nil {
		return errors.Wrapf(err, "failed to marshal %q value to json", key)
	}
	c.Spec.Raw = data

	return nil
}

// GetValue by key
// if key contains "." (a.b), find by the key path (if a:b:value in config.and get value)
func (c *Config) GetValue(key string) (any, error) {
	configMap := make(map[string]any)
	if err := json.Unmarshal(c.Spec.Raw, &configMap); err != nil {
		return nil, errors.WithStack(err)
	}
	// get all value
	if key == "" {
		return configMap, nil
	}
	// get value
	var result any = configMap
	for _, k := range strings.Split(key, ".") {
		r, ok := result.(map[string]any)
		if !ok {
			// cannot find value
			return nil, errors.Errorf("cannot find key: %s", key)
		}
		result = r[k]
	}

	return result, nil
}
