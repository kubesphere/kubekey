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
		// ======= semverCompare =======
		{
			name:      "semverCompare true-1",
			condition: []string{"{{ .foo | semverCompare \">=v1.21\" }}"},
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: true,
		},
		{
			name:      "semverCompare true-2",
			condition: []string{"{{  .foo | semverCompare \"v1.21\" }}"},
			variable: map[string]any{
				"foo": "v1.21",
			},
			excepted: true,
		},
		{
			name:      "semverCompare true-3",
			condition: []string{"{{ semverCompare \">=v1.21\" .foo }}"},
			variable: map[string]any{
				"foo": "v1.23",
			},
			excepted: true,
		},
		{
			name:      "semverCompare true-3",
			condition: []string{"{{ semverCompare \"<v1.21\" .foo }}"},
			variable: map[string]any{
				"foo": "v1.20",
			},
			excepted: true,
		},
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
			variable:  make(map[string]any),
			excepted:  true,
		},
		{
			name:      "eq true-1",
			condition: []string{"{{ and .foo (ne .foo \"\") }}"},
			variable:  make(map[string]any),
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
			variable:  make(map[string]any),
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
		// ======= hasPrefix =======
		{
			name:      "hasPrefix true-1",
			condition: []string{`{{ .foo | hasPrefix "version.BuildInfo" }}`},
			variable: map[string]any{
				"foo": `version.BuildInfo{Version:"v3.14.3", GitCommit:"f03cc04caaa8f6d7c3e67cf918929150cf6f3f12", GitTreeState:"clean", GoVersion:"go1.22.1"}`,
				"bar": "v3.14.3",
			},
			excepted: true,
		},
		// ======= empty =======
		{
			name:      "empty true-1",
			condition: []string{`{{ empty .foo }}`},
			variable: map[string]any{
				"foo": map[string]any{},
			},
			excepted: true,
		},
		{
			name:      "empty true-2",
			condition: []string{`{{ empty .foo }}`},
			variable: map[string]any{
				"foo": []any{},
			},
			excepted: true,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := ParseBool(tc.variable, tc.condition...)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.excepted, b)
		})
	}
}

func TestParseValue(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable map[string]any
		excepted []byte
	}{
		{
			name:  "single level",
			input: "{{ .foo }}",
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: []byte("bar"),
		},
		{
			name:  "multi level 1",
			input: "{{ get .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: []byte("bar"),
		},
		{
			name:  "multi level 2",
			input: "{{ get .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: []byte("bar"),
		},
		{
			name:  "multi level 2",
			input: "{{ index .foo \"foo\" }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "bar",
				},
			},
			excepted: []byte("bar"),
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
			excepted: []byte("bar"),
		},
		{
			name:  "exist value",
			input: "{{ if .foo }}{{ .foo }}{{ end }}",
			variable: map[string]any{
				"foo": "bar",
			},
			excepted: []byte("bar"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := Parse(tc.variable, tc.input)
			assert.Equal(t, tc.excepted, output)
		})
	}
}

func TestParseFunction(t *testing.T) {
	testcases := []struct {
		name     string
		input    string
		variable map[string]any
		excepted []byte
	}{
		// ======= if =======
		{
			name:  "if map 1",
			input: "{{ if .foo.foo.foo1 | eq \"bar1\" }}{{ $.foo.foo.foo2 }}{{ end }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			excepted: []byte("bar2"),
		},
		{
			name:  "if map 1",
			input: "{{ if .foo.foo.foo1 | eq \"bar1\" }}{{ .foo.foo.foo2 }}{{ end }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			excepted: []byte("bar2"),
		},
		// ======= range =======
		{
			name:  "range map 1",
			input: "{{ range $k,$v := .foo.foo }}{{ $v }}{{ end }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			excepted: []byte("bar1bar2"),
		},
		{
			name:  "range map value 1",
			input: "{{ range .foo }}{{ .foo1 }}{{ end }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			excepted: []byte("bar1"),
		},
		{
			name:  "range map top-value 1",
			input: "{{ range $_ := .foo }}{{ $.foo1 }}{{ end }}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
				"foo1": "bar11",
			},
			excepted: []byte("bar11"),
		},
		{
			name:  "range slice value 1",
			input: "{{ range .foo }}{{ .foo1 }}{{ end }}",
			variable: map[string]any{
				"foo": []map[string]any{
					{
						"foo1": "bar1",
						"foo2": "bar2",
					},
				},
			},
			excepted: []byte("bar1"),
		},
		{
			name:  "range slice value 1",
			input: "{{ range .foo }}{{ . }}{{ end }}",
			variable: map[string]any{
				"foo": []string{
					"foo1", "bar1",
				},
			},
			excepted: []byte("foo1bar1"),
		},
		// ======= default =======
		{
			name:     "default string 1",
			input:    "{{ .foo | default \"bar\" }}",
			variable: make(map[string]any),
			excepted: []byte("bar"),
		},
		{
			name:     "default string 2",
			input:    "{{ default .foo \"bar\" }}",
			variable: make(map[string]any),
			excepted: []byte("bar"),
		},

		{
			name:     "default number 1",
			input:    "{{ .foo | default 1 }}",
			variable: make(map[string]any),
			excepted: []byte("1"),
		},
		// ======= split =======
		{
			name:  "split 1",
			input: "{{ split \",\" .foo }}",
			variable: map[string]any{
				"foo": "a,b",
			},
			excepted: []byte("map[_0:a _1:b]"),
		},
		{
			name:  "split 2",
			input: "{{ .foo | split \",\" }}",
			variable: map[string]any{
				"foo": "a,b",
			},
			excepted: []byte("map[_0:a _1:b]"),
		},
		// ======= len =======
		{
			name:  "len 1",
			input: "{{ len .foo  }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("2"),
		},
		{
			name:  "len 2",
			input: "{{ .foo | len }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("2"),
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
			excepted: []byte("a"),
		},
		{
			name:  "index 2",
			input: "{{ if index .foo \"a\" }}true{{else}}false{{end}}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": "a",
				},
			},
			excepted: []byte("false"),
		},
		{
			name:  "index 3",
			input: "{{ index .foo \"foo\" \"a\"}}",
			variable: map[string]any{
				"foo": map[string]any{
					"foo": map[string]any{
						"a": "b",
					},
				},
			},
			excepted: []byte("b"),
		},
		// ======= first =======
		{
			name:  "first 1",
			input: "{{ .foo | first }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("a"),
		},
		{
			name:  "first 2",
			input: "{{ first .foo }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("a"),
		},
		// ======= last =======
		{
			name:  "last 1",
			input: "{{ .foo | last }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("b"),
		},
		{
			name:  "last 2",
			input: "{{ last .foo }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("b"),
		},
		// ======= slice =======
		{
			name:  "slice 1",
			input: "{{ slice .foo 0 2 }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("[a b]"),
		},
		// ======= join =======
		{
			name:  "join 1",
			input: "{{ slice .foo 0 2 | join \".\" }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("a.b"),
		},
		// ======= toJson =======
		{
			name:  "toJson 1",
			input: "{{ .foo | toJson }}",
			variable: map[string]any{
				"foo": []string{"a", "b"},
			},
			excepted: []byte("[\"a\",\"b\"]"),
		},
		// ======= indent =======
		{
			name:  "indent 1",
			input: "{{ .foo | indent 2 }}",
			variable: map[string]any{
				"foo": "a1: b1\na2: b2",
			},
			excepted: []byte("  a1: b1\n  a2: b2"),
		},
		// ======= printf =======
		{
			name:  "printf 1",
			input: "{{ printf \"http://%s\" .foo }}",
			variable: map[string]any{
				"foo": "a",
			},
			excepted: []byte("http://a"),
		},
		{
			name:  "printf 2",
			input: "{{ .foo | printf \"http://%s\" }}",
			variable: map[string]any{
				"foo": "a",
			},
			excepted: []byte("http://a"),
		},

		// ======= div =======
		{
			name:  "div 1",
			input: "{{ mod .foo 2 }}",
			variable: map[string]any{
				"foo": 5,
			},
			excepted: []byte("1"),
		},
		{
			name:  "div 1",
			input: "{{ mod .foo 2 }}",
			variable: map[string]any{
				"foo": 4,
			},
			excepted: []byte("0"),
		},
		// ======= sub =======
		{
			name:  "sub 1",
			input: "{{ sub .foo 2 }}",
			variable: map[string]any{
				"foo": 5,
			},
			excepted: []byte("3"),
		},
		// ======= trimPrefix =======
		{
			name:  "trimPrefix 1",
			input: `{{ .foo | trimPrefix "v" }}`,
			variable: map[string]any{
				"foo": "v1.1",
			},
			excepted: []byte("1.1"),
		},
		{
			name:     "trimPrefix 2",
			input:    `{{ .foo | default "" |trimPrefix "v" }}`,
			variable: make(map[string]any),
			excepted: nil,
		},
		// ======= fromJson =======
		{
			name:  "fromJson 1",
			input: `{{ .foo | fromJson | first }}`,
			variable: map[string]any{
				"foo": "[\"a\",\"b\"]",
			},
			excepted: []byte("a"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := Parse(tc.variable, tc.input)
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
		// ======= fromYaml =======
		{
			name:  "fromYaml 1",
			input: "{{ .foo | fromYaml | toJson }}",
			variable: map[string]any{
				"foo": `
a1: b1
a2:
  b2: 1`,
			},
			excepted: "{\"a1\":\"b1\",\"a2\":{\"b2\":1}}",
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
		// ======= ipFamily =======
		{
			name:     "ipFamily for ip address",
			input:    `{{ .ip_addr | default "10.233.64.0/18" | splitList "," | first | ipFamily }}`,
			variable: make(map[string]any),
			excepted: "IPv4",
		},
		{
			name:     "ipFamily for ip cidr",
			input:    `{{ .ip_cidr | default "10.233.64.0/18" | splitList "," | first | ipFamily }}`,
			variable: make(map[string]any),
			excepted: "IPv4",
		},
		// ======= pow =======
		{
			name:     "pow true-1",
			input:    "{{ pow 2 3 }}",
			variable: make(map[string]any),
			excepted: "8",
		},
		// ======= subtractList =======
		{
			name:     "subtractList true-1",
			input:    `{{ subtractList (list 1 2 3 4)  (list 2 4) }}`,
			variable: make(map[string]any),
			excepted: "[1 3]",
		},
		{
			name:  "subtractList true-2",
			input: `{{ subtractList .list1 .list2 }}`,
			variable: map[string]any{
				"list1": []any{1, 2, 3, 4},
				"list2": []any{2, 4},
			},
			excepted: "[1 3]",
		},
		{
			name:     "subtractList empty result",
			input:    `{{ subtractList (list 1 2)  (list 1 2) }}`,
			variable: make(map[string]any),
			excepted: "[]",
		},
		{
			name:     "subtractList with empty second list",
			input:    `{{ subtractList (list 1 2 3)  (list) }}`,
			variable: make(map[string]any),
			excepted: "[1 2 3]",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := ParseFunc(tc.variable, tc.input, func(b []byte) string { return string(b) })
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.excepted, output)
		})
	}
}
