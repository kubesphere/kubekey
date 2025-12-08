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

	"github.com/kubesphere/kubekey/v4/pkg/utils"
)

func TestCombineSlice(t *testing.T) {
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
			ac := CombineSlice(tc.g1, tc.g2)
			assert.Equal(t, tc.except, ac)
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
			assert.ElementsMatch(t, tc.except, hostsInGroup(tc.inventory, tc.groupName, utils.NewKahnGraph()))
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

func TestPrintVar(t *testing.T) {
	testcases := []struct {
		name     string
		ctx      map[string]any
		keys     []string
		excepted any
	}{
		{
			name: "strings value",
			ctx: map[string]any{
				"msg": "a",
			},
			keys:     []string{"msg"},
			excepted: "a",
		},
		{
			name: "int value",
			ctx: map[string]any{
				"msg": 1,
			},
			keys:     []string{"msg"},
			excepted: 1,
		},
		{
			name: "float value",
			ctx: map[string]any{
				"msg": 1.1,
			},
			keys:     []string{"msg"},
			excepted: 1.1,
		},
		{
			name: "map value",
			ctx: map[string]any{
				"msg": map[string]any{
					"a1": "b1",
				},
			},
			keys: []string{"msg"},
			excepted: map[string]any{
				"a1": "b1",
			},
		},
		{
			name: "slice value",
			ctx: map[string]any{
				"msg": []any{"a1", 1},
			},
			keys:     []string{"msg"},
			excepted: []any{"a1", 1},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			actual, _ := PrintVar(tc.ctx, tc.keys...)
			assert.Equal(t, tc.excepted, actual)
		})
	}
}
