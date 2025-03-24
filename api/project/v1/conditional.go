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
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
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
func (w *When) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if IsTmplSyntax(node.Value) {
			w.Data = []string{node.Value}
		} else {
			w.Data = []string{"{{ " + node.Value + " }}"}
		}
	case yaml.SequenceNode:
		if err := node.Decode(&w.Data); err != nil {
			return errors.WithStack(err)
		}
		for i, v := range w.Data {
			if !IsTmplSyntax(v) {
				w.Data[i] = ParseTmplSyntax(v)
			}
		}
	default:
		return errors.New("unsupported type, excepted string or array of strings")
	}

	return nil
}
