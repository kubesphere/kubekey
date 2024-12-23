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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

func TestMergeVariable(t *testing.T) {
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

func TestMergeGroup(t *testing.T) {
	testcases := []struct {
		name   string
		g1     []string
		g2     []string
		except []string
	}{
		{
			name: "non-repeat",
			g1: []string{
				"h1", "h2", "h3",
			},
			g2: []string{
				"h4", "h5",
			},
			except: []string{
				"h1", "h2", "h3", "h4", "h5",
			},
		},
		{
			name: "repeat value",
			g1: []string{
				"h1", "h2", "h3",
			},
			g2: []string{
				"h3", "h4", "h5",
			},
			except: []string{
				"h1", "h2", "h3", "h4", "h5",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ac := mergeSlice(tc.g1, tc.g2)
			assert.Equal(t, tc.except, ac)
		})
	}
}

func TestParseVariable(t *testing.T) {
	testcases := []struct {
		name   string
		data   map[string]any
		base   map[string]any
		except map[string]any
	}{
		{
			name: "parse string",
			data: map[string]any{
				"a": "{{ .a }}",
			},
			base: map[string]any{
				"a": "b",
			},
			except: map[string]any{
				"a": "b",
			},
		},
		{
			name: "parse map",
			data: map[string]any{
				"a": "{{ .a.b }}",
			},
			base: map[string]any{
				"a": map[string]any{
					"b": "c",
				},
			},
			except: map[string]any{
				"a": "c",
			},
		},
		{
			name: "parse slice",
			data: map[string]any{
				"a": []string{"{{ .b }}"},
			},
			base: map[string]any{
				"b": "c",
			},
			except: map[string]any{
				"a": []string{"c"},
			},
		},
		{
			name: "parse map in slice",
			data: map[string]any{
				"a": []map[string]any{
					{
						"a1": []any{"{{ .b }}"},
					},
				},
			},
			base: map[string]any{
				"b": "c",
			},
			except: map[string]any{
				"a": []map[string]any{
					{
						"a1": []any{"c"},
					},
				},
			},
		},
		{
			name: "parse slice with bool value",
			data: map[string]any{
				"a": []any{"{{ .b }}"},
			},
			base: map[string]any{
				"b": "true",
			},
			except: map[string]any{
				"a": []any{true},
			},
		},
		{
			name: "parse map with bool value",
			data: map[string]any{
				"a": "{{ .b }}",
			},
			base: map[string]any{
				"b": "true",
			},
			except: map[string]any{
				"a": true,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseVariable(tc.data, func(s string) (string, error) {
				// parse use total variable. the task variable should not contain template syntax.
				return tmpl.ParseString(CombineVariables(tc.data, tc.base), s)
			})
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, tc.data)
		})
	}
}

func TestHostsInGroup(t *testing.T) {
	testcases := []struct {
		name      string
		inventory kkcorev1.Inventory
		groupName string
		except    []string
	}{
		{
			name: "single group",
			inventory: kkcorev1.Inventory{
				Spec: kkcorev1.InventorySpec{
					Groups: map[string]kkcorev1.InventoryGroup{
						"g1": {
							Hosts: []string{"h1", "h2", "h3"},
						},
					},
				},
			},
			groupName: "g1",
			except:    []string{"h1", "h2", "h3"},
		},
		{
			name: "group in group",
			inventory: kkcorev1.Inventory{
				Spec: kkcorev1.InventorySpec{
					Groups: map[string]kkcorev1.InventoryGroup{
						"g1": {
							Hosts:  []string{"h1", "h2", "h3"},
							Groups: []string{"g2"},
						},
						"g2": {
							Hosts: []string{"h4"},
						},
					},
				},
			},
			groupName: "g1",
			except:    []string{"h1", "h2", "h3", "h4"},
		},
		{
			name: "repeat hosts in group",
			inventory: kkcorev1.Inventory{
				Spec: kkcorev1.InventorySpec{
					Groups: map[string]kkcorev1.InventoryGroup{
						"g1": {
							Hosts:  []string{"h1", "h2", "h3"},
							Groups: []string{"g2"},
						},
						"g2": {
							Hosts: []string{"h3", "h4"},
						},
					},
				},
			},
			groupName: "g1",
			except:    []string{"h4", "h1", "h2", "h3"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.except, HostsInGroup(tc.inventory, tc.groupName))
		})
	}
}

func TestExtension2Slice(t *testing.T) {
	testcases := []struct {
		name   string
		data   map[string]any
		ext    runtime.RawExtension
		except []any
	}{
		{
			name: "succeed",
			data: map[string]any{
				"a": []any{"a1", "a2"},
			},
			ext: runtime.RawExtension{
				Raw: []byte(`{{ .a | toJson }}`),
			},
			except: []any{"a1", "a2"},
		},
		{
			name: "empty ext",
			data: map[string]any{
				"b": []any{"b1", "b2"},
			},
			ext: runtime.RawExtension{
				Raw: []byte(`{{ .a | toJson }}`),
			},
			except: make([]any, 0),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.except, Extension2Slice(tc.data, tc.ext))
		})
	}
}
