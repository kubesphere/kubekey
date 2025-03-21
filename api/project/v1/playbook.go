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
	"strings"

	"github.com/cockroachdb/errors"
)

// NOTE:
// To unmarshal into a specific field of a struct, the field name must be explicitly specified.
// Otherwise, the UnmarshalYAML method associated with the field’s struct will be treated as a method of the parent struct, rather than the field itself.

// Playbook defined in project.
type Playbook struct {
	Play []Play
}

// Validate playbook. delete empty ImportPlaybook which has convert to play.
func (p *Playbook) Validate() error {
	var newPlay = make([]Play, 0)
	for _, play := range p.Play {
		//  import_playbook is a link, should be ignored.
		if play.ImportPlaybook != "" {
			continue
		}

		if len(play.PlayHost.Hosts) == 0 {
			return errors.New("playbook's hosts must not be empty")
		}
		newPlay = append(newPlay, play)
	}
	p.Play = newPlay

	return nil
}

// IsTmplSyntax Check if the string conforms to the template syntax.
func IsTmplSyntax(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}

// ParseTmplSyntax wraps a string with template syntax delimiters "{{" and "}}"
// to make it a valid Go template expression
func ParseTmplSyntax(s string) string {
	return "{{ " + s + "}}"
}
