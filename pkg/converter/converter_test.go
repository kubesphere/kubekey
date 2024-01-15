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

package converter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
)

func TestMarshalPlaybook(t *testing.T) {
	testcases := []struct {
		name   string
		file   string
		except *kkcorev1.Playbook
	}{
		{
			name: "marshal playbook",
			file: "playbooks/playbook1.yaml",
			except: &kkcorev1.Playbook{[]kkcorev1.Play{
				{
					Base:     kkcorev1.Base{Name: "play1"},
					PlayHost: kkcorev1.PlayHost{Hosts: []string{"localhost"}},
					Roles: []kkcorev1.Role{
						{kkcorev1.RoleInfo{
							Role: "role1",
							Block: []kkcorev1.Block{
								{
									BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "role1 | block1"}},
									Task: kkcorev1.Task{UnknownFiled: map[string]any{
										"debug": map[string]any{
											"msg": "echo \"hello world\"",
										},
									}},
								},
							},
						}},
					},
					Handlers: nil,
					PreTasks: []kkcorev1.Block{
						{
							BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | pre_block1"}},
							Task: kkcorev1.Task{UnknownFiled: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
					PostTasks: []kkcorev1.Block{
						{
							BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | post_block1"}},
							Task: kkcorev1.Task{UnknownFiled: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
					Tasks: []kkcorev1.Block{
						{
							BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | block1"}},
							BlockInfo: kkcorev1.BlockInfo{Block: []kkcorev1.Block{
								{
									BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | block1 | block1"}},
									Task: kkcorev1.Task{UnknownFiled: map[string]any{
										"debug": map[string]any{
											"msg": "echo \"hello world\"",
										},
									}},
								},
								{
									BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | block1 | block2"}},
									Task: kkcorev1.Task{UnknownFiled: map[string]any{
										"debug": map[string]any{
											"msg": "echo \"hello world\"",
										},
									}},
								},
							}},
						},
						{
							BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play1 | block2"}},
							Task: kkcorev1.Task{UnknownFiled: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
				},
				{
					Base:     kkcorev1.Base{Name: "play2"},
					PlayHost: kkcorev1.PlayHost{Hosts: []string{"localhost"}},
					Tasks: []kkcorev1.Block{
						{
							BlockBase: kkcorev1.BlockBase{Base: kkcorev1.Base{Name: "play2 | block1"}},
							Task: kkcorev1.Task{UnknownFiled: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
				},
			}},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pb, err := MarshalPlaybook(os.DirFS("testdata"), tc.file)
			assert.NoError(t, err)
			assert.Equal(t, tc.except, pb)
		})
	}
}

func TestGroupHostBySerial(t *testing.T) {
	hosts := []string{"h1", "h2", "h3", "h4", "h5", "h6", "h7"}
	testcases := []struct {
		name         string
		serial       []any
		exceptResult [][]string
		exceptErr    bool
	}{
		{
			name:   "group host by 1",
			serial: []any{1},
			exceptResult: [][]string{
				{"h1"},
				{"h2"},
				{"h3"},
				{"h4"},
				{"h5"},
				{"h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 2",
			serial: []any{2},
			exceptResult: [][]string{
				{"h1", "h2"},
				{"h3", "h4"},
				{"h5", "h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  2",
			serial: []any{1, 2},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3"},
				{"h4", "h5"},
				{"h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  40%",
			serial: []any{"1", "40%"},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3", "h4"},
				{"h5", "h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:         "group host by unSupport serial type",
			serial:       []any{1.1},
			exceptResult: nil,
			exceptErr:    true,
		},
		{
			name:         "group host by unSupport serial value",
			serial:       []any{"%1.1%"},
			exceptResult: nil,
			exceptErr:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GroupHostBySerial(hosts, tc.serial)
			if tc.exceptErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.exceptResult, result)
			}
		})
	}
}
