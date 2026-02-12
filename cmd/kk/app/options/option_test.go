/*
Copyright 2024 The KubeSphere Authors.

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

package options

import (
	"encoding/json"
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
)

func TestParseKey(t *testing.T) {
	tests := []struct {
		key      string
		expected []string
	}{
		{"a.b.c", []string{"a", "b", "c"}},
		{"a[0].b", []string{"a", "0", "b"}},
		{"a[0][1].b", []string{"a", "0", "1", "b"}},
		{"array[2].field.subfield", []string{"array", "2", "field", "subfield"}},
		{"simple", []string{"simple"}},
		{"a.b[0].c[1].d", []string{"a", "b", "0", "c", "1", "d"}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := parseKey(tt.key)
			if len(result) != len(tt.expected) {
				t.Errorf("parseKey(%q) = %v, want %v", tt.key, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseKey(%q)[%d] = %q, want %q", tt.key, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestParseAndSetValue(t *testing.T) {
	tests := []struct {
		name     string
		setVal   string
		expected map[string]interface{}
	}{
		{
			name:   "simple string",
			setVal: "key=value",
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:   "nested object",
			setVal: "outer.inner=value",
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name:   "boolean true",
			setVal: "flag=true",
			expected: map[string]interface{}{
				"flag": true,
			},
		},
		{
			name:   "boolean FALSE",
			setVal: "flag=FALSE",
			expected: map[string]interface{}{
				"flag": false,
			},
		},
		{
			name:   "numeric integer",
			setVal: "count=42",
			expected: map[string]interface{}{
				"count": int64(42),
			},
		},
		{
			name:   "numeric float",
			setVal: "price=3.14",
			expected: map[string]interface{}{
				"price": 3.14,
			},
		},
		{
			name:   "array index",
			setVal: "items[0]=first",
			expected: map[string]interface{}{
				"items": []interface{}{
					"first",
				},
			},
		},
		{
			name:   "array index nested",
			setVal: "users[0].name=john",
			expected: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"name": "john",
					},
				},
			},
		},
		{
			name:   "JSON object",
			setVal: "config={\"key\":\"value\"}",
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:   "JSON array",
			setVal: "items=[1,2,3]",
			expected: map[string]interface{}{
				"items": []interface{}{
					float64(1), float64(2), float64(3),
				},
			},
		},
		{
			name:   "value with dot",
			setVal: "key=c.d",
			expected: map[string]interface{}{
				"key": "c.d",
			},
		},
		{
			name:   "value with multiple dots",
			setVal: "key=com.example.test",
			expected: map[string]interface{}{
				"key": "com.example.test",
			},
		},
		{
			name:   "value with dots in nested",
			setVal: "outer.inner=john.doe",
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "john.doe",
				},
			},
		},
		{
			name:   "value with dots in array element",
			setVal: "users[0].name=jane.doe",
			expected: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"name": "jane.doe",
					},
				},
			},
		},
		{
			name:   "escaped comma in value",
			setVal: "tags=a\\,b\\,c",
			expected: map[string]interface{}{
				"tags": "a,b,c",
			},
		},
		{
			name:   "escaped comma in nested value",
			setVal: "config.list=one\\,two\\,three",
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"list": "one,two,three",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &kkcorev1.Config{}
			err := parseAndSetValue(config, tt.setVal)
			if err != nil {
				t.Fatalf("parseAndSetValue(%q) failed: %v", tt.setVal, err)
			}

			// Get the value from config
			configVal := config.Value()

			// Compare the result
			resultJSON, _ := json.Marshal(configVal)
			expectedJSON, _ := json.Marshal(tt.expected)

			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("parseAndSetValue(%q) = %s, want %s", tt.setVal, resultJSON, expectedJSON)
			}
		})
	}
}

func TestMultipleValuesWithDots(t *testing.T) {
	tests := []struct {
		name     string
		setVals  []string
		expected map[string]interface{}
	}{
		{
			name:    "multiple nested fields with dots in values",
			setVals: []string{"a.b=c.d", "a.c=c.f"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "c.d",
					"c": "c.f",
				},
			},
		},
		{
			name:    "multiple array elements",
			setVals: []string{"a.d[0]=1", "a.d[1]=2"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"d": []interface{}{
						int64(1), int64(2),
					},
				},
			},
		},
		{
			name:    "mixed nested and array",
			setVals: []string{"a.x=foo", "a.y[0]=bar", "a.y[1]=baz"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"x": "foo",
					"y": []interface{}{
						"bar", "baz",
					},
				},
			},
		},
		{
			name:    "escaped comma in multiple values",
			setVals: []string{"a.tags=x\\,y\\,z", "a.name=test"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"tags": "x,y,z",
					"name": "test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &kkcorev1.Config{}
			for _, setVal := range tt.setVals {
				err := parseAndSetValue(config, setVal)
				if err != nil {
					t.Fatalf("parseAndSetValue(%q) failed: %v", setVal, err)
				}
			}

			configVal := config.Value()

			resultJSON, _ := json.Marshal(configVal)
			expectedJSON, _ := json.Marshal(tt.expected)

			if string(resultJSON) != string(expectedJSON) {
				t.Errorf("parseAndSetValue(%v) = %s, want %s", tt.setVals, resultJSON, expectedJSON)
			}
		})
	}
}
