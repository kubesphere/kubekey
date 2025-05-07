package project

import (
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
							Vars: yaml.Node{
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
							Vars: yaml.Node{
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
							Vars: yaml.Node{
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
						PlayHost: kkprojectv1.PlayHost{
							Hosts: []string{"node1"},
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
