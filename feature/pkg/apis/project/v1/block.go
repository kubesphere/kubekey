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

	"k8s.io/klog/v2"
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
	Delegatable      `yaml:",inline"`
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

	// UnknownField store undefined filed
	UnknownField map[string]any `yaml:"-"`
}

// UnmarshalYAML yaml string to block.
func (b *Block) UnmarshalYAML(unmarshal func(any) error) error {
	// fill baseInfo
	var bb BlockBase
	if err := unmarshal(&bb); err == nil {
		b.BlockBase = bb
	}

	var m map[string]any
	if err := unmarshal(&m); err != nil {
		klog.Errorf("unmarshal data to map error: %v", err)

		return err
	}

	if includeTasks, ok := handleIncludeTasks(m); ok {
		// Set the IncludeTasks field if "include_tasks" exists and is valid.
		b.IncludeTasks = includeTasks

		return nil
	}

	switch {
	case m["block"] != nil:
		// If the "block" key exists, unmarshal it into BlockInfo and set the BlockInfo field.
		bi, err := handleBlock(m, unmarshal)
		if err != nil {
			return err
		}
		b.BlockInfo = bi
	default:
		// If neither "include_tasks" nor "block" are present, treat the data as a task.
		t, err := handleTask(m, unmarshal)
		if err != nil {
			return err
		}
		b.Task = t
		// Set any remaining unknown fields to the Task's UnknownField.
		b.UnknownField = m
	}

	return nil
}

// handleIncludeTasks checks if the "include_tasks" key exists in the map and is of type string.
// If so, it returns the string value and true, otherwise it returns an empty string and false.
func handleIncludeTasks(m map[string]any) (string, bool) {
	if v, ok := m["include_tasks"]; ok {
		if it, ok := v.(string); ok {
			return it, true
		}
	}

	return "", false
}

// handleBlock attempts to unmarshal the block data into a BlockInfo structure.
// If successful, it returns the BlockInfo and nil. If an error occurs, it logs the error and returns it.
func handleBlock(_ map[string]any, unmarshal func(any) error) (BlockInfo, error) {
	var bi BlockInfo
	if err := unmarshal(&bi); err != nil {
		klog.Errorf("unmarshal data to block error: %v", err)

		return bi, err
	}

	return bi, nil
}

// handleTask attempts to unmarshal the task data into a Task structure.
// If successful, it deletes existing fields from the map, logs the error if it occurs, and returns the Task and nil.
func handleTask(m map[string]any, unmarshal func(any) error) (Task, error) {
	var t Task
	if err := unmarshal(&t); err != nil {
		klog.Errorf("unmarshal data to task error: %v", err)

		return t, err
	}
	deleteExistField(reflect.TypeOf(Block{}), m)

	return t, nil
}

func deleteExistField(rt reflect.Type, m map[string]any) {
	for i := range rt.NumField() {
		field := rt.Field(i)
		if field.Anonymous {
			deleteExistField(field.Type, m)
		} else {
			if isFound := deleteField(rt.Field(i), m); isFound {
				break
			}
		}
	}
}

// deleteField find and delete the filed, return the field if found.
func deleteField(field reflect.StructField, m map[string]any) bool {
	yamlTag := field.Tag.Get("yaml")
	if yamlTag != "" {
		for _, t := range strings.Split(yamlTag, ",") {
			if _, ok := m[t]; ok {
				delete(m, t)

				return true
			}
		}
	} else {
		t := strings.ToUpper(field.Name[:1]) + field.Name[1:]
		if _, ok := m[t]; ok {
			delete(m, t)

			return true
		}
	}

	return false
}
