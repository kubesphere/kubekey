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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"reflect"
	"strings"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Namespaced

type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              runtime.RawExtension `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}

func (c *Config) SetValue(key string, value any) error {
	configMap := make(map[string]any)
	if err := json.Unmarshal(c.Spec.Raw, &configMap); err != nil {
		return err
	}
	// set value
	var f func(input map[string]any, key []string, value any) any
	f = func(input map[string]any, key []string, value any) any {
		if len(key) == 1 {
			input[key[0]] = value
		} else if len(key) > 1 {
			if v, ok := input[key[0]]; ok && reflect.TypeOf(v).Kind() == reflect.Map {
				input[key[0]] = f(v.(map[string]any), key[1:], value)
			} else {
				input[key[0]] = f(make(map[string]any), key[1:], value)
			}
		}
		return input
	}
	data, err := json.Marshal(f(configMap, strings.Split(key, "."), value))
	if err != nil {
		return err
	}
	c.Spec.Raw = data
	return nil
}

func (c *Config) GetValue(key string) (any, error) {
	configMap := make(map[string]any)
	if err := json.Unmarshal(c.Spec.Raw, &configMap); err != nil {
		return nil, err
	}
	// get value
	return configMap[key], nil
}
