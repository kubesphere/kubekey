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

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
)

func TestGetAllVariable(t *testing.T) {
	testcases := []struct {
		name   string
		value  value
		except map[string]any
	}{
		{
			name: "global override runtime variable",
			value: value{
				Config: kubekeyv1.Config{
					Spec: runtime.RawExtension{
						Raw: []byte(`
artifact:
  images:
   - abc
`),
					},
				},
				Inventory: kubekeyv1.Inventory{},
				Hosts: map[string]host{
					"test": {
						RuntimeVars: map[string]any{
							"artifact": map[string]any{
								"k1": "v1",
								"k2": 2,
								"k3": true,
								"k4": map[string]any{
									"k41": "v41",
								},
							},
						},
					},
				},
			},
			except: map[string]any{
				"artifact": map[string]any{
					"k1": "v1",
					"k2": 2,
					"k3": true,
					"k4": map[string]any{
						"k41": "v41",
					},
					"images": []any{"abc"},
				},
				"groups": map[string]interface{}{"all": []string{"localhost"}},
				"inventory_hosts": map[string]interface{}{
					"test": map[string]interface{}{
						"artifact": map[string]interface{}{
							"images": []interface{}{"abc"},
						},
						"inventory_name": "test",
					},
				},
				"inventory_name": "test",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := variable{value: &tc.value}
			result, err := v.Get(GetAllVariable("test"))
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, result)
		})
	}
}
