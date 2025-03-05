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
	"gopkg.in/yaml.v3"
)

// Role defined in project.
type Role struct {
	RoleInfo
}

// RoleInfo defined in project.
type RoleInfo struct {
	Base             `yaml:",inline"`
	Conditional      `yaml:",inline"`
	Taggable         `yaml:",inline"`
	CollectionSearch `yaml:",inline"`

	// Role ref in playbook
	Role string `yaml:"role,omitempty"`

	Block []Block
}

// UnmarshalYAML yaml string to role.
func (r *Role) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		r.Role = node.Value
	case yaml.MappingNode:
		return node.Decode(&r.RoleInfo)
	}

	return nil
}
