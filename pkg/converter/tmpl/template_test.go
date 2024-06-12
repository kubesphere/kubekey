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

	"github.com/flosch/pongo2/v6"
	"github.com/stretchr/testify/assert"
)

func TestParseBool(t *testing.T) {
	testcases := []struct {
		name      string
		condition []string
		variable  pongo2.Context
		excepted  bool
	}{
		{
			name:      "parse success",
			condition: []string{"foo == \"bar\""},
			variable: pongo2.Context{
				"foo": "bar",
			},
			excepted: true,
		},
		{
			name:      "in array",
			condition: []string{"test in inArr"},
			variable: pongo2.Context{
				"test":  "a",
				"inArr": []string{"a", "b"},
			},
			excepted: true,
		},
		{
			name:      "container string",
			condition: []string{"instr[0].test"},
			variable: pongo2.Context{
				"instr": []pongo2.Context{
					{"test": true},
					{"test": false},
				},
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
		variable pongo2.Context
		excepted string
	}{
		{
			name:  "parse success",
			input: "{{foo}}",
			variable: pongo2.Context{
				"foo": "bar",
			},
			excepted: "bar",
		},
		{
			name:  "parse in map",
			input: "{% for _,v in value %}{{v.a}}{% endfor %}",
			variable: pongo2.Context{
				"value": pongo2.Context{
					"foo": pongo2.Context{
						"a": "b",
					},
				},
			},
			excepted: "b",
		},
		{
			name:  "parse in",
			input: "{% set k=value['foo'] %}{{ k.a }}",
			variable: pongo2.Context{
				"value": pongo2.Context{
					"foo": pongo2.Context{
						"a": "b",
					},
				},
			},
			excepted: "b",
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
		variable pongo2.Context
		excepted string
	}{
		{
			name: "parse success",
			variable: pongo2.Context{
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
