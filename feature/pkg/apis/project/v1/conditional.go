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
	"errors"
)

// Conditional defined in project.
type Conditional struct {
	When When `yaml:"when,omitempty"`
}

// When defined in project.
type When struct {
	Data []string
}

// UnmarshalYAML yaml string to when
func (w *When) UnmarshalYAML(unmarshal func(any) error) error {
	var s string
	if err := unmarshal(&s); err == nil {
		w.Data = []string{s}

		return nil
	}

	var a []string
	if err := unmarshal(&a); err == nil {
		w.Data = a

		return nil
	}

	return errors.New("unsupported type, excepted string or array of strings")
}
