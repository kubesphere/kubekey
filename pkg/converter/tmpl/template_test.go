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

package tmpl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

func TestParseBool(t *testing.T) {
	testcases := []struct {
		name      string
		condition []string
		variable  variable.VariableData
		excepted  bool
	}{
		{
			name:      "parse success",
			condition: []string{"foo == \"bar\""},
			variable: variable.VariableData{
				"foo": "bar",
			},
			excepted: true,
		},
		{
			name:      "in",
			condition: []string{"test in inArr"},
			variable: variable.VariableData{
				"test":  "a",
				"inArr": []string{"a", "b"},
			},
			excepted: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := ParseBool(tc.variable, tc.condition)
			assert.Equal(t, tc.excepted, b)
		})
	}
}

func TestParseString(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable variable.VariableData
		excepted string
	}{
		{
			name:  "parse success",
			input: "{{foo}}",
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: "bar",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := ParseString(tc.variable, tc.input)
			assert.Equal(t, tc.excepted, output)
		})
	}
}

func TestParseFile(t *testing.T) {
	testcases := []struct {
		name     string
		variable variable.VariableData
		excepted string
	}{
		{
			name: "parse success",
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: "foo: bar",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := ParseFile(tc.variable, []byte("foo: {{foo}}"))
			assert.Equal(t, tc.excepted, output)
		})
	}
}
