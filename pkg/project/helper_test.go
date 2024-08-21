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

package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	kkprojectv1 "github.com/kubesphere/kubekey/v4/pkg/apis/project/v1"
)

func TestGetPlaybookBaseFromAbsPlaybook(t *testing.T) {
	testcases := []struct {
		name         string
		basePlaybook string
		playbook     string
		except       string
	}{
		{
			name:         "find from project/playbooks/playbook",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook2.yaml",
			except:       filepath.Join("playbooks", "playbook2.yaml"),
		},
		{
			name:         "find from current_playbook/playbooks/playbook",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook3.yaml",
			except:       filepath.Join("playbooks", "playbooks", "playbook3.yaml"),
		},
		{
			name:         "cannot find",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			playbook:     "playbook4.yaml",
			except:       "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, getPlaybookBaseFromPlaybook(os.DirFS("testdata"), tc.basePlaybook, tc.playbook))
		})
	}
}

func TestGetRoleBaseFromAbsPlaybook(t *testing.T) {
	testcases := []struct {
		name         string
		basePlaybook string
		roleName     string
		except       string
	}{
		{
			name:         "find from project/roles/roleName",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role1",
			except:       filepath.Join("roles", "role1"),
		},
		{
			name:         "find from current_playbook/roles/roleName",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role2",
			except:       filepath.Join("playbooks", "roles", "role2"),
		},
		{
			name:         "cannot find",
			basePlaybook: filepath.Join("playbooks", "playbook1.yaml"),
			roleName:     "role3",
			except:       "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, getRoleBaseFromPlaybook(os.DirFS("testdata"), tc.basePlaybook, tc.roleName))
		})
	}
}

func TestGetYamlFile(t *testing.T) {
	testcases := []struct {
		name   string
		base   string
		except string
	}{
		{
			name:   "get yaml",
			base:   filepath.Join("playbooks", "playbook2"),
			except: filepath.Join("playbooks", "playbook2.yaml"),
		},
		{
			name:   "get yml",
			base:   filepath.Join("playbooks", "playbook3"),
			except: filepath.Join("playbooks", "playbook3.yml"),
		},
		{
			name:   "cannot find",
			base:   filepath.Join("playbooks", "playbook4"),
			except: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.except, getYamlFile(os.DirFS("testdata"), tc.base))
		})
	}
}

func TestMarshalPlaybook(t *testing.T) {
	testcases := []struct {
		name   string
		file   string
		except *kkprojectv1.Playbook
	}{
		{
			name: "marshal playbook",
			file: "playbooks/playbook1.yaml",
			except: &kkprojectv1.Playbook{Play: []kkprojectv1.Play{
				{
					Base:     kkprojectv1.Base{Name: "play1"},
					PlayHost: kkprojectv1.PlayHost{Hosts: []string{"localhost"}},
					Roles: []kkprojectv1.Role{
						{
							RoleInfo: kkprojectv1.RoleInfo{
								Role: "role1",
								Block: []kkprojectv1.Block{
									{
										BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "role1 | block1"}},
										Task: kkprojectv1.Task{UnknownField: map[string]any{
											"debug": map[string]any{
												"msg": "echo \"hello world\"",
											},
										}},
									},
								},
							},
						},
					},
					Handlers: nil,
					PreTasks: []kkprojectv1.Block{
						{
							BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | pre_block1"}},
							Task: kkprojectv1.Task{UnknownField: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
					PostTasks: []kkprojectv1.Block{
						{
							BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | post_block1"}},
							Task: kkprojectv1.Task{UnknownField: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
					Tasks: []kkprojectv1.Block{
						{
							BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | block1"}},
							BlockInfo: kkprojectv1.BlockInfo{Block: []kkprojectv1.Block{
								{
									BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | block1 | block1"}},
									Task: kkprojectv1.Task{UnknownField: map[string]any{
										"debug": map[string]any{
											"msg": "echo \"hello world\"",
										},
									}},
								},
								{
									BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | block1 | block2"}},
									Task: kkprojectv1.Task{UnknownField: map[string]any{
										"debug": map[string]any{
											"msg": "echo \"hello world\"",
										},
									}},
								},
							}},
						},
						{
							BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play1 | block2"}},
							Task: kkprojectv1.Task{UnknownField: map[string]any{
								"debug": map[string]any{
									"msg": "echo \"hello world\"",
								},
							}},
						},
					},
				},
				{
					Base:     kkprojectv1.Base{Name: "play2"},
					PlayHost: kkprojectv1.PlayHost{Hosts: []string{"localhost"}},
					Tasks: []kkprojectv1.Block{
						{
							BlockBase: kkprojectv1.BlockBase{Base: kkprojectv1.Base{Name: "play2 | block1"}},
							Task: kkprojectv1.Task{UnknownField: map[string]any{
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
			pb, err := marshalPlaybook(os.DirFS("testdata"), tc.file)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, pb)
		})
	}
}

func TestCombineMaps(t *testing.T) {
	testcases := []struct {
		name   string
		v1     map[string]any
		v2     map[string]any
		except map[string]any
		err    bool
	}{
		{
			name: "v1 is null",
			v2: map[string]any{
				"a": "b",
			},
			except: map[string]any{
				"a": "b",
			},
		},
		{
			name: "success",
			v1: map[string]any{
				"a1": "b1",
			},
			v2: map[string]any{
				"a2": "b2",
			},
			except: map[string]any{
				"a1": "b1",
				"a2": "b2",
			},
		},
		{
			name: "duplicate key",
			v1: map[string]any{
				"a1": "b1",
			},
			v2: map[string]any{
				"a1": "b2",
			},
			err: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			maps, err := combineMaps(tc.v1, tc.v2)
			if tc.err {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.except, maps)
			}
		})
	}
}
