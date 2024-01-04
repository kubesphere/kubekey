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

import "fmt"

type Play struct {
	ImportPlaybook string `yaml:"import_playbook,omitempty"`

	Base             `yaml:",inline"`
	Taggable         `yaml:",inline"`
	CollectionSearch `yaml:",inline"`

	PlayHost PlayHost `yaml:"hosts,omitempty"`

	// Facts
	GatherFacts bool `yaml:"gather_facts,omitempty"`

	// defaults to be deprecated, should be 'None' in future
	//GatherSubset []GatherSubset
	//GatherTimeout int
	//FactPath string

	// Variable Attribute
	VarsFiles  []string `yaml:"vars_files,omitempty"`
	VarsPrompt []string `yaml:"vars_prompt,omitempty"`

	// Role Attributes
	Roles []Role `yaml:"roles,omitempty"`

	// Block (Task) Lists Attributes
	Handlers  []Block `yaml:"handlers,omitempty"`
	PreTasks  []Block `yaml:"pre_tasks,omitempty"`
	PostTasks []Block `yaml:"post_tasks,omitempty"`
	Tasks     []Block `yaml:"tasks,omitempty"`

	// Flag/Setting Attributes
	ForceHandlers     bool       `yaml:"force_handlers,omitempty"`
	MaxFailPercentage float32    `yaml:"percent,omitempty"`
	Serial            PlaySerial `yaml:"serial,omitempty"`
	Strategy          string     `yaml:"strategy,omitempty"`
	Order             string     `yaml:"order,omitempty"`
}

type PlaySerial struct {
	Data []any
}

func (s *PlaySerial) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var as []any
	if err := unmarshal(&as); err == nil {
		s.Data = as
		return nil
	}
	var a any
	if err := unmarshal(&a); err == nil {
		s.Data = []any{a}
		return nil
	}
	return fmt.Errorf("unsupported type, excepted any or array")

}

type PlayHost struct {
	Hosts []string
}

func (p *PlayHost) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var hs []string
	if err := unmarshal(&hs); err == nil {
		p.Hosts = hs
		return nil
	}
	var h string
	if err := unmarshal(&h); err == nil {
		p.Hosts = []string{h}
		return nil
	}
	return fmt.Errorf("unsupported type, excepted string or string array")

}
