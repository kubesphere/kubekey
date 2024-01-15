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
)

func TestMergeVariable(t *testing.T) {
	testcases := []struct {
		name     string
		v1       VariableData
		v2       VariableData
		excepted VariableData
	}{
		{
			name: "primary variables value is empty",
			v1:   nil,
			v2: VariableData{
				"a1": "v1",
			},
			excepted: VariableData{
				"a1": "v1",
			},
		},
		{
			name: "auxiliary variables value is empty",
			v1: VariableData{
				"p1": "v1",
			},
			v2: nil,
			excepted: VariableData{
				"p1": "v1",
			},
		},
		{
			name: "non-repeat value",
			v1: VariableData{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
				},
			},
			v2: VariableData{
				"a1": "v1",
			},
			excepted: VariableData{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
				},
				"a1": "v1",
			},
		},
		{
			name: "repeat value",
			v1: VariableData{
				"p1": "v1",
				"p2": map[string]any{
					"p21": "v21",
					"p22": "v22",
				},
			},
			v2: VariableData{
				"a1": "v1",
				"p1": "v2",
				"p2": map[string]any{
					"p21": "v22",
					"a21": "v21",
				},
			},
			excepted: VariableData{
				"a1": "v1",
				"p1": "v2",
				"p2": map[string]any{
					"p21": "v22",
					"a21": "v21",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v := mergeVariables(tc.v1, tc.v2)
			assert.Equal(t, tc.excepted, v)
		})
	}
}
