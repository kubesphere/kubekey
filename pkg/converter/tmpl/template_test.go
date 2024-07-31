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
)

func TestParseBool(t *testing.T) {
	testcases := []struct {
		name      string
		condition []string
		variable  map[string]any
		excepted  bool
	}{
		// ======= eq =======
		{
			name:      "atoi true-1",
			condition: []string{`{{ .foo | trimSuffix " kB" | atoi | le .bar }}`},
			variable: map[string]any{
				"foo": "8148172 kB",
				"bar": 10,
			},
			excepted: true,
		},
		// ======= eq =======
		{
			name:      "eq true-1",
			condition: []string{"{{ eq .foo \"bar\" }}"},
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: true,
		},
		// ======= ne =======
		{
			name:      "eq true-1",
			condition: []string{"{{ ne .foo \"\" }}"},
			variable:  map[string]any{},
			excepted:  true,
		},
		{
			name:      "eq true-1",
			condition: []string{"{{ and .foo (ne .foo \"\") }}"},
			variable:  map[string]any{},
			excepted:  false,
		},
		// ======= value exist =======
		{
			name:      "value exist true-1",
			condition: []string{"{{ .foo }}"},
			variable: map[string]any{
				"foo": "true",
			},
			excepted: true,
		},
		{
			name:      "value exist false-1",
			condition: []string{"{{ .foo }}"},
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: false,
		},
		{
			name:      "value exist false-2",
			condition: []string{"{{ .foo }}"},
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: false,
		},
		// ======= default =======
		{
			name:      "default true-1",
			condition: []string{"{{ .foo | default true }}"},
			variable:  map[string]any{},
			excepted:  true,
		},
		// ======= has =======
		{
			name:      "has true-1",
			condition: []string{"{{ .foo | has \"a\" }}"},
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: true,
		},
		// ======= regexMatch =======
		{
			name:      "regexMatch true-1",
			condition: []string{`{{ regexMatch "^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])|(([0-9a-fA-F]{1,4}:){7}([0-9a-fA-F]{1,4}|:)|(([0-9a-fA-F]{1,4}:){1,6}|:):([0-9a-fA-F]{1,4}|:){1,6}([0-9a-fA-F]{1,4}|:)))$" .foo }}`},
			variable: map[string]any{
				"foo": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			},
			excepted: true,
		},
		{
			name:      "regexMatch true-2",
			condition: []string{`{{ regexMatch "^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\\.(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])|(([0-9a-fA-F]{1,4}:){7}([0-9a-fA-F]{1,4}|:)|(([0-9a-fA-F]{1,4}:){1,6}|:):([0-9a-fA-F]{1,4}|:){1,6}([0-9a-fA-F]{1,4}|:)))$" .foo }}`},
			variable: map[string]any{
				"foo": "1.1.1.1",
			},
			excepted: true,
		},
		{
			name:      "regexMatch true-3",
			condition: []string{`{{ regexMatch "^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$" .foo }}`},
			variable: map[string]any{
				"foo": "a.b",
			},
			excepted: true,
		},
		{
			name:      "regexMatch false-1",
			condition: []string{`{{ regexMatch "^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$" .foo }}`},
			variable: map[string]any{
				"foo": "a.=b",
			},
			excepted: false,
		},
		// ======= contains =======
		{
			name:      "contains true-1",
			condition: []string{`{{ .foo | contains (printf "Version:\"%s\"" .bar) }}`},
			variable: map[string]any{
				"foo": `version.BuildInfo{Version:"v3.14.3", GitCommit:"f03cc04caaa8f6d7c3e67cf918929150cf6f3f12", GitTreeState:"clean", GoVersion:"go1.22.1"}`,
				"bar": "v3.14.3",
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

func TestParseValue(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable map[string]any
		excepted string
	}{
		{
			name:  "single level",
			input: "{{ .foo }}",
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: "bar",
		},
		{
			name:  "multi level 1",
			input: "{{ get .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: "bar",
		},
		{
			name:  "multi level 2",
			input: "{{ get .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: "bar",
		},
		{
			name:  "multi level 2",
			input: "{{ index .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: "bar",
		},
		{
			name:  "multi level 3",
			input: "{{ index .foo \"foo\" \"foo\" \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo": map[string]any{
							"foo": "bar",
						},
					},
				},
			},
			excepted: "bar",
		},
		{
			name:  "exist value",
			input: "{{ if .foo }}{{ .foo }}{{ end }}",
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

func TestParseFunction(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable map[string]any
		excepted string
	}{
		// ======= default =======
		{
			name:     "default string 1",
			input:    "{{ .foo | default \"bar\" }}",
			variable: map[string]any{},
			excepted: "bar",
		},
		{
			name:     "default string 2",
			input:    "{{ default .foo \"bar\" }}",
			variable: map[string]any{},
			excepted: "bar",
		},

		{
			name:     "default number 1",
			input:    "{{ .foo | default 1 }}",
			variable: map[string]any{},
			excepted: "1",
		},
		// ======= split =======
		{
			name:  "split 1",
			input: "{{ split \",\" .foo }}",
			variable: map[string]any{
				"foo": "a,b",
			},
			excepted: "map[_0:a _1:b]",
		},
		{
			name:  "split 2",
			input: "{{ .foo | split \",\" }}",
			variable: map[string]any{
				"foo": "a,b",
			},
			excepted: "map[_0:a _1:b]",
		},
		// ======= len =======
		{
			name:  "len 1",
			input: "{{ len .foo  }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "2",
		},
		{
			name:  "len 2",
			input: "{{ .foo | len }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "2",
		},
		// ======= index =======
		{
			name:  "index 1",
			input: "{{ index .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "a",
				},
			},
			excepted: "a",
		},
		{
			name:  "index 1",
			input: "{{ if index .foo \"a\" }}true{{else}}false{{end}}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "a",
				},
			},
			excepted: "false",
		},
		// ======= first =======
		{
			name:  "first 1",
			input: "{{ .foo | first }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "a",
		},
		{
			name:  "first 2",
			input: "{{ first .foo }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "a",
		},
		// ======= last =======
		{
			name:  "last 1",
			input: "{{ .foo | last }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "b",
		},
		{
			name:  "last 2",
			input: "{{ last .foo }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "b",
		},
		// ======= slice =======
		{
			name:  "slice 1",
			input: "{{ slice .foo 0 2 }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "[a b]",
		},
		// ======= join =======
		{
			name:  "join 1",
			input: "{{ slice .foo 0 2 | join \".\" }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "a.b",
		},
		// ======= toJson =======
		{
			name:  "toJson 1",
			input: "{{ .foo | toJson }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: "[\"a\",\"b\"]",
		},
		// ======= toYaml =======
		{
			name:  "toYaml 1",
			input: "{{ .foo | toYaml }}",
			variable: map[string]any{
				"foo": map[string]any{
					"a1": "b1",
					"a2": "b2",
				},
			},
			excepted: "a1: b1\na2: b2",
		},
		// ======= indent =======
		{
			name:  "indent 1",
			input: "{{ .foo | indent 2 }}",
			variable: map[string]any{
				"foo": "a1: b1\na2: b2",
			},
			excepted: "  a1: b1\n  a2: b2",
		},
		// ======= printf =======
		{
			name:  "printf 1",
			input: "{{ printf \"http://%s\" .foo }}",
			variable: map[string]any{
				"foo": "a",
			},
			excepted: "http://a",
		},
		{
			name:  "printf 2",
			input: "{{ .foo | printf \"http://%s\" }}",
			variable: map[string]any{
				"foo": "a",
			},
			excepted: "http://a",
		},

		// ======= div =======
		{
			name:  "div 1",
			input: "{{ mod .foo 2 }}",
			variable: map[string]any{
				"foo": 5,
			},
			excepted: "1",
		},
		{
			name:  "div 1",
			input: "{{ mod .foo 2 }}",
			variable: map[string]any{
				"foo": 4,
			},
			excepted: "0",
		},
		// ======= sub =======
		{
			name:  "sub 1",
			input: "{{ sub .foo 2 }}",
			variable: map[string]any{
				"foo": 5,
			},
			excepted: "3",
		},
		// ======= trimPrefix =======
		{
			name:  "trimPrefix 1",
			input: `{{ .foo | trimPrefix "v" }}`,
			variable: map[string]any{
				"foo": "v1.1",
			},
			excepted: "1.1",
		},
		{
			name:     "trimPrefix 2",
			input:    `{{ .foo | default "" |trimPrefix "v" }}`,
			variable: map[string]any{},
			excepted: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ParseString(tc.variable, tc.input)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.excepted, output)
		})
	}
}

func TestParseCustomFunction(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable map[string]any
		excepted string
	}{
		// ======= versionAtLeast =======
		{
			name:  "versionAtLeast true-1",
			input: "{{ .foo | versionAtLeast \"v1.21\" }}",
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: "true",
		},
		{
			name:  "versionAtLeast true-2",
			input: "{{  .foo | versionAtLeast \"v1.21\" }}",
			variable: map[string]any{
				"foo": "v1.21",
			},
			excepted: "true",
		},
		{
			name:  "versionAtLeast true-3",
			input: "{{ versionAtLeast \"v1.21\" .foo }}",
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: "true",
		},
		// ======= versionLessThan =======
		{
			name:  "versionLessThan true-1",
			input: "{{ versionLessThan \"v1.25\" .foo }}",
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: "true",
		},
		{
			name:  "versionLessThan true-2",
			input: "{{  .foo | versionLessThan \"v1.25\" }}",
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: "true",
		},
		// ======= ipInCIDR =======
		{
			name:  "ipInCIDR true-1",
			input: "{{ ipInCIDR 0 .foo }}",
			variable: map[string]any{
				"foo": "10.233.0.0/18",
			},
			excepted: "10.233.0.1",
		},
		{
			name:  "ipInCIDR true-2",
			input: "{{ .foo | ipInCIDR 0 }}",
			variable: map[string]any{
				"foo": "10.233.0.0/18",
			},
			excepted: "10.233.0.1",
		},
		{
			name:  "ipInCIDR true-3",
			input: "{{ ipInCIDR -1 .foo }}",
			variable: map[string]any{
				"foo": "10.233.0.0/18",
			},
			excepted: "10.233.63.254",
		},
		// ======= pow =======
		{
			name:     "pow true-1",
			input:    "{{ pow 2 3 }}",
			variable: map[string]any{},
			excepted: "8",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ParseString(tc.variable, tc.input)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.excepted, output)
		})
	}
}
