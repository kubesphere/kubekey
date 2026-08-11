package project

import (
	"encoding/json"
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestMarshalPlaybook(t *testing.T) {
	testcases := []struct {
		name     string
		playbook kkcorev1.Playbook
		except   *kkprojectv1.Playbook
	}{
		{
			name: "test_playbook1",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook1.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "playbook1",
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
						PreTasks: []kkprojectv1.Block{
							{
								BlockBase: kkprojectv1.BlockBase{
									Base: kkprojectv1.Base{
										Name: "task1",
									},
								},
								Task: kkprojectv1.Task{
									UnknownField: map[string]any{
										"annotations": map[string]string{
											kkcorev1alpha1.TaskAnnotationRelativePath: ".",
										},
										"debug": map[string]any{
											"msg": "im task1",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test_playbook2",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook2.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "playbook2",
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
						Roles: []kkprojectv1.Role{
							{
								RoleInfo: kkprojectv1.RoleInfo{
									Role: "role1",
									Block: []kkprojectv1.Block{
										{
											BlockBase: kkprojectv1.BlockBase{
												Base: kkprojectv1.Base{
													Name: "task1",
												},
											},
											Task: kkprojectv1.Task{
												UnknownField: map[string]any{
													"annotations": map[string]string{
														kkcorev1alpha1.TaskAnnotationRelativePath: "roles/role1",
													},
													"debug": map[string]any{
														"msg": "im task1",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test_playbook3",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook3.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "playbook3",
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
						Tasks: []kkprojectv1.Block{
							{
								IncludeTasks: "include_task1.yaml",
								BlockInfo: kkprojectv1.BlockInfo{
									Block: []kkprojectv1.Block{
										{
											BlockBase: kkprojectv1.BlockBase{
												Base: kkprojectv1.Base{
													Name: "task1",
												},
											},
											Task: kkprojectv1.Task{
												UnknownField: map[string]any{
													"annotations": map[string]string{
														kkcorev1alpha1.TaskAnnotationRelativePath: ".",
													},
													"debug": map[string]any{
														"msg": "im task1",
													},
												},
											},
										},
										{
											IncludeTasks: "include_task1_1.yaml",
											BlockInfo: kkprojectv1.BlockInfo{
												Block: []kkprojectv1.Block{
													{
														BlockBase: kkprojectv1.BlockBase{
															Base: kkprojectv1.Base{
																Name: "task2",
															},
														},
														Task: kkprojectv1.Task{
															UnknownField: map[string]any{
																"annotations": map[string]string{
																	kkcorev1alpha1.TaskAnnotationRelativePath: ".",
																},
																"debug": map[string]any{
																	"msg": "im task2",
																},
															},
														},
													},
												},
											},
										},
										{
											IncludeTasks: "include_task1/include_task1_2.yaml",
											BlockInfo: kkprojectv1.BlockInfo{
												Block: []kkprojectv1.Block{
													{
														BlockBase: kkprojectv1.BlockBase{
															Base: kkprojectv1.Base{
																Name: "task3",
															},
														},
														Task: kkprojectv1.Task{
															UnknownField: map[string]any{
																"annotations": map[string]string{
																	kkcorev1alpha1.TaskAnnotationRelativePath: ".",
																},
																"debug": map[string]any{
																	"msg": "im task3",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test_playbook4",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbooks/playbook4.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "playbook4_1",
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
						Roles: []kkprojectv1.Role{
							{
								RoleInfo: kkprojectv1.RoleInfo{
									Role: "role2",
									Block: []kkprojectv1.Block{
										{
											IncludeTasks: "include_task1/include_task1.yaml",
											BlockInfo: kkprojectv1.BlockInfo{
												Block: []kkprojectv1.Block{
													{
														IncludeTasks: "include_task2.yaml",
														BlockInfo: kkprojectv1.BlockInfo{
															Block: []kkprojectv1.Block{
																{
																	BlockBase: kkprojectv1.BlockBase{
																		Base: kkprojectv1.Base{
																			Name: "task2",
																		},
																	},
																	Task: kkprojectv1.Task{
																		UnknownField: map[string]any{
																			"annotations": map[string]string{
																				kkcorev1alpha1.TaskAnnotationRelativePath: "roles/role2",
																			},
																			"debug": map[string]any{
																				"msg": "im task2",
																			},
																		},
																	},
																},
															},
														},
													},
													{
														IncludeTasks: "include_task3.yaml",
														BlockInfo: kkprojectv1.BlockInfo{
															Block: []kkprojectv1.Block{
																{
																	BlockBase: kkprojectv1.BlockBase{
																		Base: kkprojectv1.Base{
																			Name: "task3",
																		},
																	},
																	Task: kkprojectv1.Task{
																		UnknownField: map[string]any{
																			"annotations": map[string]string{
																				kkcorev1alpha1.TaskAnnotationRelativePath: "roles/role2",
																			},
																			"debug": map[string]any{
																				"msg": "im task3",
																			},
																		},
																	},
																},
															},
														},
													},
													{
														BlockBase: kkprojectv1.BlockBase{
															Base: kkprojectv1.Base{
																Name: "task1",
															},
														},
														Task: kkprojectv1.Task{
															UnknownField: map[string]any{
																"annotations": map[string]string{
																	kkcorev1alpha1.TaskAnnotationRelativePath: "roles/role2",
																},
																"debug": map[string]any{
																	"msg": "im task1",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "test_vars_1",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook_var1.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "playbook-var1",
							Vars: kkprojectv1.Vars{
								Nodes: []yaml.Node{
									{
										Kind:   yaml.MappingNode,
										Tag:    "!!map",
										Line:   6,
										Column: 5,
										Content: []*yaml.Node{
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a",
												Line:   6,
												Column: 5,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "b",
												Line:   6,
												Column: 8,
											},
										},
									},
								},
							},
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
					},
				},
			},
		},
		{
			name: "test_vars_2",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook_var2.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						VarsFiles: []string{"vars/var1.yaml", "vars/var2.yaml"},
						Base: kkprojectv1.Base{
							Name: "playbook-var2",
							Vars: kkprojectv1.Vars{
								Nodes: []yaml.Node{
									{
										Kind:   yaml.MappingNode,
										Tag:    "!!map",
										Line:   2,
										Column: 1,
										Content: []*yaml.Node{
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a1",
												Line:   2,
												Column: 1,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "aa",
												Line:   2,
												Column: 5,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a2",
												Line:   3,
												Column: 1,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!int",
												Value:  "1",
												Line:   3,
												Column: 5,
											},
										},
									},
									{
										Kind:   yaml.MappingNode,
										Tag:    "!!map",
										Line:   1,
										Column: 1,
										Content: []*yaml.Node{
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a2",
												Line:   1,
												Column: 1,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "aaa",
												Line:   1,
												Column: 5,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a3",
												Line:   2,
												Column: 1,
											},
											{
												Kind:   yaml.MappingNode,
												Tag:    "!!map",
												Value:  "",
												Line:   3,
												Column: 2,
												Content: []*yaml.Node{
													{
														Kind:   yaml.ScalarNode,
														Tag:    "!!str",
														Value:  "b3",
														Line:   3,
														Column: 2,
													},
													{
														Kind:   yaml.ScalarNode,
														Tag:    "!!int",
														Value:  "1",
														Line:   3,
														Column: 6,
													},
												},
											},
										},
									},
								},
							},
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
					},
				},
			},
		},
		{
			name: "test_vars_3",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbook_var3.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						VarsFiles: []string{"vars/var1.yaml"},
						Base: kkprojectv1.Base{
							Name: "playbook-var3",
							Vars: kkprojectv1.Vars{
								Nodes: []yaml.Node{
									{
										Kind:   yaml.MappingNode,
										Tag:    "!!map",
										Line:   8,
										Column: 5,
										Content: []*yaml.Node{
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a2",
												Line:   8,
												Column: 5,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!int",
												Value:  "2",
												Line:   8,
												Column: 9,
											},
										},
									},
									{
										Kind:   yaml.MappingNode,
										Tag:    "!!map",
										Line:   2,
										Column: 1,
										Content: []*yaml.Node{
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a1",
												Line:   2,
												Column: 1,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "aa",
												Line:   2,
												Column: 5,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!str",
												Value:  "a2",
												Line:   3,
												Column: 1,
											},
											{
												Kind:   yaml.ScalarNode,
												Tag:    "!!int",
												Value:  "1",
												Line:   3,
												Column: 5,
											},
										},
									},
								},
							},
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
					},
				},
			},
		},
		{
			name: "test_playbook5_dependency_role",
			playbook: kkcorev1.Playbook{
				Spec: kkcorev1.PlaybookSpec{
					Playbook: "testdata/playbooks/playbook5.yaml",
				},
			},
			except: &kkprojectv1.Playbook{
				Play: []kkprojectv1.Play{
					{
						Base: kkprojectv1.Base{
							Name: "dependency role",
						},
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
						},
						Roles: []kkprojectv1.Role{
							{
								RoleInfo: kkprojectv1.RoleInfo{
									Role: "role3",
									RoleDependency: []kkprojectv1.Role{
										{
											RoleInfo: kkprojectv1.RoleInfo{
												Role: "role1",
												Block: []kkprojectv1.Block{
													{
														BlockBase: kkprojectv1.BlockBase{
															Base: kkprojectv1.Base{
																Name: "task1",
															},
														},
														Task: kkprojectv1.Task{
															UnknownField: map[string]any{
																"annotations": map[string]string{
																	kkcorev1alpha1.TaskAnnotationRelativePath: "roles/role1",
																},
																"debug": map[string]any{
																	"msg": "im task1",
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			project, err := newLocalProject(tc.playbook)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := project.MarshalPlaybook()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, actual)
		})
	}
}

func TestInjectPlaybooksOrder(t *testing.T) {
	// Source order_base.yaml has plays a(0), b(1), c(2).
	// Configured injections: d(-1), e(-1.1), f(-1), g(0).
	// Sort contract:
	//   e(-1.1) < d/f(-1, config seq d<f) < g(0, config) < a(0, file) < b(1) < c(2)
	// => e, d, f, g, a, b, c
	playbook := kkcorev1.Playbook{
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "testdata/playbooks/order_base.yaml",
			Config: mustConfig(t, map[string]any{
				"playbooks": []any{
					map[string]any{"order": -1.0, "path": "order_d.yaml"},
					map[string]any{"order": -1.1, "path": "order_e.yaml"},
					map[string]any{"order": -1.0, "path": "order_f.yaml"},
					map[string]any{"order": 0.0, "path": "order_g.yaml"},
				},
			}),
		},
	}
	project, err := newLocalProject(playbook)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := project.MarshalPlaybook()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range actual.Play {
		got = append(got, p.Name)
	}
	assert.Equal(t, []string{"e", "d", "f", "g", "a", "b", "c"}, got)
}

func TestInjectPlaybooksOrderWithImportDirective(t *testing.T) {
	// Source order_src.yaml:
	//   [0] import_playbook: order_pre.yaml  (pure import directive -> weight 0)
	//   [1] play a                          (weight 1)
	//   [2] play b                          (weight 2)
	// order_pre.yaml loads a single play "pre".
	// Inject x at order 0 (same weight as the directive; config wins) =>
	// final = [x, pre, a, b]. This proves that a pure import_playbook directive
	// still occupies an order position, so config injections can be positioned
	// relative to it.
	playbook := kkcorev1.Playbook{
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "testdata/playbooks/order_src.yaml",
			Config: mustConfig(t, map[string]any{
				"playbooks": []any{
					map[string]any{"order": 0.0, "path": "order_x.yaml"},
				},
			}),
		},
	}
	project, err := newLocalProject(playbook)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := project.MarshalPlaybook()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range actual.Play {
		got = append(got, p.Name)
	}
	assert.Equal(t, []string{"x", "pre", "a", "b"}, got)
}

func TestInjectPlaybooksOrderEmptyPath(t *testing.T) {
	// Source order_base.yaml: a(0), b(1), c(2).
	// An injection whose path renders to empty must be skipped silently; a
	// configured path keeps its position. Here:
	//   order 0.5 -> empty path  -> skipped
	//   order 1.5 -> order_d.yaml -> inserted between b and c
	// => a, b, d, c
	playbook := kkcorev1.Playbook{
		Spec: kkcorev1.PlaybookSpec{
			Playbook: "testdata/playbooks/order_base.yaml",
			Config: mustConfig(t, map[string]any{
				"playbooks": []any{
					map[string]any{"order": 0.5, "path": "{{ .not_set | default \"\" }}"},
					map[string]any{"order": 1.5, "path": "order_d.yaml"},
				},
			}),
		},
	}
	project, err := newLocalProject(playbook)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := project.MarshalPlaybook()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range actual.Play {
		got = append(got, p.Name)
	}
	assert.Equal(t, []string{"a", "b", "d", "c"}, got)
}

// mustConfig builds a Config from a map by going through JSON unmarshalling so
// that Spec.Object (used by Value()) is populated. The map is wrapped under
// "spec" to match the Config.Spec (runtime.RawExtension) JSON shape.
func mustConfig(t *testing.T, v map[string]any) kkcorev1.Config {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"spec": v})
	if err != nil {
		t.Fatal(err)
	}
	var cfg kkcorev1.Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	return cfg
}
