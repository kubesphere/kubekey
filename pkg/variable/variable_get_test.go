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

package variable

import (
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetHostnames(t *testing.T) {
	testcases := []struct {
		name     string
		hosts    []string
		variable Variable
		except   []string
	}{
		{
			name:  "host value",
			hosts: []string{"n1"},
			variable: &variable{
				value: &value{
					Inventory: kkcorev1.Inventory{
						Spec: kkcorev1.InventorySpec{
							Hosts: map[string]runtime.RawExtension{
								"node1": {},
								"node2": {},
							},
						},
					},
					Hosts: map[string]host{
						"n1": {},
						"n2": {},
					},
				},
			},
			except: []string{"n1"},
		},
		{
			name:  "group value",
			hosts: []string{"g1"},
			variable: &variable{
				value: &value{
					Inventory: kkcorev1.Inventory{
						Spec: kkcorev1.InventorySpec{
							Hosts: map[string]runtime.RawExtension{
								"n1": {},
								"n2": {},
							},
							Groups: map[string]kkcorev1.InventoryGroup{
								"g1": {
									Hosts: []string{"n1"},
								},
							},
						},
					},
					Hosts: map[string]host{
						"n1": {},
						"n2": {},
					},
				},
			},
			except: []string{"n1"},
		},
		{
			name:  "group index value",
			hosts: []string{"g1[0]"},
			variable: &variable{
				value: &value{
					Inventory: kkcorev1.Inventory{
						Spec: kkcorev1.InventorySpec{
							Hosts: map[string]runtime.RawExtension{
								"n1": {},
								"n2": {},
							},
							Groups: map[string]kkcorev1.InventoryGroup{
								"g1": {
									Hosts: []string{"n1", "n2"},
								},
							},
						},
					},
					Hosts: map[string]host{
						"n1": {},
						"n2": {},
					},
				},
			},
			except: []string{"n1"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.variable.Get(GetHostnames(tc.hosts))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, result)
		})
	}
}

func TestGetAllVariable(t *testing.T) {
	testcases := []struct {
		name     string
		variable Variable
		except   map[string]any
	}{
		{
			name: "global override runtime variable",
			variable: &variable{
				value: &value{
					Config: kkcorev1.Config{
						Spec: runtime.RawExtension{
							Raw: []byte(`{
"artifact": {
  "images": [ "abc" ]
}
}`)},
					},
					Inventory: kkcorev1.Inventory{
						Spec: kkcorev1.InventorySpec{
							Hosts: map[string]runtime.RawExtension{
								"localhost": {Raw: []byte(`{
"internal_ipv4": "127.0.0.1",
"internal_ipv6": "::1"
}`)},
							},
						},
					},
					Hosts: map[string]host{
						"localhost": {},
					},
				},
			},
			except: map[string]any{
				"internal_ipv4": "127.0.0.1",
				"internal_ipv6": "::1",
				"artifact": map[string]any{
					"images": []any{"abc"},
				},
				"groups": map[string][]string{"all": {"localhost"}, "ungrouped": {"localhost"}},
				"inventory_hosts": map[string]any{
					"localhost": map[string]any{
						"internal_ipv4": "127.0.0.1",
						"internal_ipv6": "::1",
						"artifact": map[string]any{
							"images": []any{"abc"},
						},
						"inventory_name": "localhost",
						"hostname":       "localhost",
					},
				},
				"inventory_name": "localhost",
				"hostname":       "localhost",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.variable.Get(GetAllVariable("localhost"))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, result)
		})
	}
}

func TestGetHostMaxLength(t *testing.T) {
	testcases := []struct {
		name     string
		variable Variable
		except   int
	}{
		{
			name: "length",
			variable: &variable{
				value: &value{
					Hosts: map[string]host{
						"n1":  {},
						"n22": {},
					},
				},
			},
			except: 3,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.variable.Get(GetHostMaxLength())
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, result)
		})
	}
}

func TestGetWorkdir(t *testing.T) {
	testcases := []struct {
		name     string
		variable Variable
		except   string
	}{
		{
			name: "workdir",
			variable: &variable{
				value: &value{
					Config: kkcorev1.Config{
						Spec: runtime.RawExtension{
							Raw: []byte("{\"work_dir\": \"abc\"}"),
							Object: &unstructured.Unstructured{Object: map[string]any{
								"work_dir": "abc",
							}},
						},
					},
				},
			},
			except: "abc",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.variable.Get(GetWorkDir())
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, result)
		})
	}
}

func TestCombineVariables(t *testing.T) {
	testcases := []struct {
		name     string
		v1       map[string]any
		v2       map[string]any
		excepted map[string]any
	}{
		{
			name: "primary variables value is empty",
			v1:   nil,
			v2: map[string]any{
				"a1": "v1",
			},
			excepted: map[string]any{
				"a1": "v1",
			},
		},
		{
			name: "auxiliary variables value is empty",
			v1: map[string]any{
				"p1": "v1",
			},
			v2: nil,
			excepted: map[string]any{
				"p1": "v1",
			},
		},
		{
			name: "non-repeat value",
			v1: map[string]any{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
				},
			},
			v2: map[string]any{
				"a1": "v1",
			},
			excepted: map[string]any{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
				},
				"a1": "v1",
			},
		},
		{
			name: "repeat value",
			v1: map[string]any{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
					"p22": "v22",
				},
			},
			v2: map[string]any{
				"a1": "v1",
				"p1": "v2",
				"p2": map[string]any{
					"p21": "v22",
					"a21": "v21",
				},
			},
			excepted: map[string]any{
				"a1": "v1",
				"p1": "v2",
				"p2": map[string]any{
					"p21": "v22",
					"a21": "v21",
					"p22": "v22",
				},
			},
		},
		{
			name: "repeat deep value",
			v1: map[string]any{
				"p1": map[string]string{
					"p11": "v11",
				},
				"p2": map[string]any{
					"p21": "v21",
					"p22": "v22",
				},
			},
			v2: map[string]any{
				"p1": map[string]string{
					"p21": "v21",
				},
				"p2": map[string]any{
					"p21": map[string]any{
						"p211": "v211",
					},
					"a21": "v21",
				},
			},
			excepted: map[string]any{
				"p1": map[string]any{
					"p11": "v11",
					"p21": "v21",
				},
				"p2": map[string]any{
					"p21": map[string]any{
						"p211": "v211",
					},
					"p22": "v22",
					"a21": "v21",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := CombineVariables(tc.v1, tc.v2)
			assert.Equal(t, tc.excepted, v)
		})
	}
}
