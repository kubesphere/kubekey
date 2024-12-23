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
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// Block defined in project.
type Block struct {
	BlockBase
	// If it has Block, Task should be empty
	Task
	IncludeTasks string `yaml:"include_tasks,omitempty"`

	BlockInfo
}

// BlockBase defined in project.
type BlockBase struct {
	Base             `yaml:",inline"`
	Conditional      `yaml:",inline"`
	CollectionSearch `yaml:",inline"`
	Taggable         `yaml:",inline"`
	Notifiable       `yaml:",inline"`
	Delegable        `yaml:",inline"`
}

// BlockInfo defined in project.
type BlockInfo struct {
	Block  []Block `yaml:"block,omitempty"`
	Rescue []Block `yaml:"rescue,omitempty"`
	Always []Block `yaml:"always,omitempty"`
}

// Task defined in project.
type Task struct {
	AsyncVal    int         `yaml:"async,omitempty"`
	ChangedWhen When        `yaml:"changed_when,omitempty"`
	Delay       int         `yaml:"delay,omitempty"`
	FailedWhen  When        `yaml:"failed_when,omitempty"`
	Loop        any         `yaml:"loop,omitempty"`
	LoopControl LoopControl `yaml:"loop_control,omitempty"`
	Poll        int         `yaml:"poll,omitempty"`
	Register    string      `yaml:"register,omitempty"`
	Retries     int         `yaml:"retries,omitempty"`
	Until       When        `yaml:"until,omitempty"`

	// deprecated, used to be loop and loop_args but loop has been repurposed
	//LoopWith string	`yaml:"loop_with"`

	// UnknownField store undefined field
	UnknownField map[string]any `yaml:"-"`
}

// UnmarshalYAML yaml to block.
func (b *Block) UnmarshalYAML(node *yaml.Node) error {
	// fill baseInfo
	if err := node.Decode(&b.BlockBase); err != nil {
		return fmt.Errorf("failed to decode block, error: %w", err)
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		switch keyNode.Value {
		case "include_tasks":
			b.IncludeTasks = valueNode.Value
			return nil

		case "block":
			return node.Decode(&b.BlockInfo)
		}
	}

	if err := node.Decode(&b.Task); err != nil {
		return fmt.Errorf("failed to decode task: %w", err)
	}
	b.UnknownField = collectUnknownFields(node, append(getFieldNames(reflect.TypeOf(BlockBase{})), getFieldNames(reflect.TypeOf(Task{}))...))

	return nil
}

// collectUnknownFields traverses a YAML node and collects fields that are not in the excludeFields list.
// It returns a map where the keys are the names of the unknown fields and the values are their corresponding values.
func collectUnknownFields(node *yaml.Node, excludeFields []string) map[string]any {
	unknown := make(map[string]any)
	excludeSet := make(map[string]struct{}, len(excludeFields))
	for _, field := range excludeFields {
		excludeSet[field] = struct{}{}
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if _, excluded := excludeSet[keyNode.Value]; excluded {
			continue
		}

		var value any
		if err := valueNode.Decode(&value); err == nil {
			unknown[keyNode.Value] = value
		} else {
			unknown[keyNode.Value] = fmt.Sprintf("failed to decode: %v", err)
		}
	}

	return unknown
}

// getFieldNames returns a slice of field names for a given struct type.
// It inspects the struct fields and extracts the names from the "yaml" tags.
// If a field has an "inline" tag, it recursively processes the fields of the embedded struct.
func getFieldNames(t reflect.Type) []string {
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" {
			if strings.Contains(yamlTag, "inline") {
				inlineFields := getFieldNames(field.Type)
				fields = append(fields, inlineFields...)
				continue
			}
			tagName := strings.Split(yamlTag, ",")[0]
			if tagName != "" && tagName != "-" {
				fields = append(fields, tagName)
			}
		}
	}

	return fields
}
