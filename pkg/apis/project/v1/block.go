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

type Block struct {
	BlockBase
	// If has Block, Task should be empty
	Task
	IncludeTasks string `yaml:"include_tasks,omitempty"`

	BlockInfo
}

type BlockBase struct {
	Base             `yaml:",inline"`
	Conditional      `yaml:",inline"`
	CollectionSearch `yaml:",inline"`
	Taggable         `yaml:",inline"`
	Notifiable       `yaml:",inline"`
	Delegatable      `yaml:",inline"`
}

type BlockInfo struct {
	Block  []Block `yaml:"block,omitempty"`
	Rescue []Block `yaml:"rescue,omitempty"`
	Always []Block `yaml:"always,omitempty"`
}

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

	//
	UnknownFiled map[string]any `yaml:"-"`
}

func (b *Block) UnmarshalYAML(unmarshal func(interface{}) error) error {
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

	if v, ok := m["include_tasks"]; ok {
		b.IncludeTasks = v.(string)
	} else if _, ok := m["block"]; ok {
		// render block
		var bi BlockInfo
		err := unmarshal(&bi)
		if err != nil {
			klog.Errorf("unmarshal data to block error: %v", err)
			return err
		}
		b.BlockInfo = bi
	} else {
		// render task
		var t Task
		err := unmarshal(&t)
		if err != nil {
			klog.Errorf("unmarshal data to task error: %v", err)
			return err
		}
		b.Task = t
		deleteExistField(reflect.TypeOf(Block{}), m)
		// set unknown flied to task.UnknownFiled
		b.UnknownFiled = m
	}

	return nil
}

func deleteExistField(rt reflect.Type, m map[string]any) {
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Anonymous {
			deleteExistField(field.Type, m)
		} else {
			yamlTag := rt.Field(i).Tag.Get("yaml")
			if yamlTag != "" {
				for _, t := range strings.Split(yamlTag, ",") {
					if _, ok := m[t]; ok {
						delete(m, t)
						break
					}
				}
			} else {
				t := strings.ToUpper(rt.Field(i).Name[:1]) + rt.Field(i).Name[1:]
				if _, ok := m[t]; ok {
					delete(m, t)
					break
				}
			}
		}
	}
}
