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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestValidate(t *testing.T) {
	testcases := []struct {
		name     string
		playbook Playbook
	}{
		{
			name: "host is empty",
			playbook: Playbook{Play: []Play{
				{
					Base: Base{
						Name: "test",
					},
				},
			}},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Error(t, tc.playbook.Validate())
		})
	}
}

func TestUnmarshalYamlPlaybook(t *testing.T) {
	testcases := []struct {
		name     string
		data     []byte
		excepted []Play
	}{
		{
			name: "Unmarshal hosts with single value",
			data: []byte(`---
- name: test play
  hosts: localhost
`),
			excepted: []Play{
				{
					Base:     Base{Name: "test play"},
					PlayHost: PlayHost{[]string{"localhost"}},
				},
			},
		},
		{
			name: "Unmarshal hosts with multiple value",
			data: []byte(`---
- name: test play
  hosts: ["control-plane", "worker"]
`),
			excepted: []Play{
				{
					Base: Base{
						Name: "test play",
					},
					PlayHost: PlayHost{[]string{"control-plane", "worker"}},
				},
			},
		},
		{
			name: "Unmarshal role with single value",
			data: []byte(`---
- name: test play
  hosts: localhost
  roles:
    - test
`),
			excepted: []Play{
				{
					Base: Base{Name: "test play"},
					PlayHost: PlayHost{
						[]string{"localhost"},
					},
					Roles: []Role{
						{
							RoleInfo{
								Role: "test",
							},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal role with map value",
			data: []byte(`---
- name: test play
  hosts: localhost
  roles:
    - role: test
`),
			excepted: []Play{
				{
					Base: Base{Name: "test play"},
					PlayHost: PlayHost{
						[]string{"localhost"},
					},
					Roles: []Role{
						{
							RoleInfo{
								Role: "test",
							},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal when with single value",
			data: []byte(`---
- name: test play
  hosts: localhost
  roles:
  - role: test
    when: "true"
`),
			excepted: []Play{
				{
					Base: Base{Name: "test play"},
					PlayHost: PlayHost{
						[]string{"localhost"},
					},
					Roles: []Role{
						{
							RoleInfo{
								Conditional: Conditional{When: When{Data: []string{"true"}}},
								Role:        "test",
							},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal when with multiple value",
			data: []byte(`---
- name: test play
  hosts: localhost
  roles:
  - role: test
    when: ["true","false"]
`),
			excepted: []Play{
				{
					Base: Base{Name: "test play"},
					PlayHost: PlayHost{
						[]string{"localhost"},
					},
					Roles: []Role{
						{
							RoleInfo{
								Conditional: Conditional{When: When{Data: []string{"true", "false"}}},
								Role:        "test",
							},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal single level block",
			data: []byte(`---
- name: test play
  hosts: localhost
  tasks:
    - name: test
      custom-module: abc
`),
			excepted: []Play{
				{
					Base:     Base{Name: "test play"},
					PlayHost: PlayHost{Hosts: []string{"localhost"}},
					Tasks: []Block{
						{
							BlockBase: BlockBase{Base: Base{Name: "test"}},
							Task:      Task{UnknownField: map[string]any{"custom-module": "abc"}},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal multi level block",
			data: []byte(`---
- name: test play
  hosts: localhost
  tasks:
    - name: test
      block:
      - name: test | test
        custom-module: abc
`),
			excepted: []Play{
				{
					Base:     Base{Name: "test play"},
					PlayHost: PlayHost{Hosts: []string{"localhost"}},
					Tasks: []Block{
						{
							BlockBase: BlockBase{Base: Base{Name: "test"}},
							BlockInfo: BlockInfo{
								Block: []Block{{
									BlockBase: BlockBase{Base: Base{Name: "test | test"}},
									Task:      Task{UnknownField: map[string]any{"custom-module": "abc"}},
								}},
							},
						},
					},
				},
			},
		},
		{
			name: "Unmarshal include_tasks block",
			data: []byte(`---
- name: test play
  hosts: localhost
  tasks:
    - include_tasks: task.yaml
`),
			excepted: []Play{
				{
					Base:     Base{Name: "test play"},
					PlayHost: PlayHost{Hosts: []string{"localhost"}},
					Tasks: []Block{
						{
							IncludeTasks: "task.yaml",
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var pb Playbook
			err := yaml.Unmarshal(tc.data, &pb.Play)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.excepted, pb.Play)
		})
	}
}
