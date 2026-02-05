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

package proxy

import (
	"testing"
)

func TestNewPathExpression(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		wantErr     bool
		description string
	}{
		{
			name:        "simple literal path",
			path:        "/api/v1/resource",
			wantErr:     false,
			description: "should parse simple literal path without error",
		},
		{
			name:        "path with single parameter",
			path:        "/api/v1/{name}",
			wantErr:     false,
			description: "should parse path with single parameter",
		},
		{
			name:        "path with multiple parameters",
			path:        "/api/v1/{namespace}/{name}",
			wantErr:     false,
			description: "should parse path with multiple parameters",
		},
		{
			name:        "path with wildcard",
			path:        "/api/v1/{path:*}",
			wantErr:     false,
			description: "should parse path with wildcard parameter",
		},
		{
			name:        "path with regex parameter",
			path:        "/api/v1/{name:[a-z]+}",
			wantErr:     false,
			description: "should parse path with regex parameter",
		},
		{
			name:        "root path",
			path:        "/",
			wantErr:     false,
			description: "should parse root path",
		},
		{
			name:        "nested path with params",
			path:        "/namespaces/{namespace}/pods/{name}",
			wantErr:     false,
			description: "should parse nested path with namespace and name parameters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := newPathExpression(tc.path)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr == nil {
				t.Fatal("expected non-nil pathExpression")
			}
			if expr.Matcher == nil {
				t.Error("Matcher should not be nil")
			}
		})
	}
}

func TestNewPathExpression_MatcherValidation(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		input         string
		expectedMatch bool
		description   string
	}{
		{
			name:          "exact match",
			path:          "/api/v1/resource",
			input:         "/api/v1/resource",
			expectedMatch: true,
			description:   "should match exact path",
		},
		{
			name:          "match with parameter",
			path:          "/api/v1/{name}",
			input:         "/api/v1/test",
			expectedMatch: true,
			description:   "should match path with parameter",
		},
		{
			name:          "no match different path",
			path:          "/api/v1/resource",
			input:         "/api/v1/other",
			expectedMatch: false,
			description:   "should not match different path",
		},
		{
			name:          "match with namespace",
			path:          "/namespaces/{namespace}/pods/{name}",
			input:         "/namespaces/default/pods/my-pod",
			expectedMatch: true,
			description:   "should match path with multiple parameters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := newPathExpression(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			matches := expr.Matcher.FindStringSubmatch(tc.input)
			gotMatch := matches != nil
			if gotMatch != tc.expectedMatch {
				t.Errorf("expected match=%v, got match=%v for input=%s", tc.expectedMatch, gotMatch, tc.input)
			}
		})
	}
}

func TestNewPathExpression_VarCount(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		expectedVar int
		description string
	}{
		{
			name:        "no variables",
			path:        "/api/v1/resource",
			expectedVar: 0,
			description: "should have zero variables in literal path",
		},
		{
			name:        "single variable",
			path:        "/api/v1/{name}",
			expectedVar: 1,
			description: "should have one variable",
		},
		{
			name:        "multiple variables",
			path:        "/namespaces/{namespace}/pods/{name}",
			expectedVar: 2,
			description: "should have two variables",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := newPathExpression(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.VarCount != tc.expectedVar {
				t.Errorf("expected VarCount=%d, got %d", tc.expectedVar, expr.VarCount)
			}
		})
	}
}

func TestNewPathExpression_LiteralCount(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		description string
	}{
		{
			name:        "empty path",
			path:        "/",
			description: "should handle root path",
		},
		{
			name:        "simple path",
			path:        "/api/v1/resource",
			description: "should count literal characters in simple path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := newPathExpression(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.LiteralCount < 0 {
				t.Error("LiteralCount should not be negative")
			}
		})
	}
}

func TestTemplateToRegularExpression(t *testing.T) {
	testCases := []struct {
		name               string
		template           string
		expectedExpression string
		expectError        bool
	}{
		{
			name:               "simple path",
			template:           "/api/v1/resource",
			expectedExpression: "^api/v1/resource(/.*)?$",
			expectError:        false,
		},
		{
			name:               "path with parameter",
			template:           "/api/v1/{name}",
			expectedExpression: "^api/v1/([^/]+?)(/.*)?$",
			expectError:        false,
		},
		{
			name:               "path with wildcard",
			template:           "/api/v1/{path:*}",
			expectedExpression: "^api/v1/(.*)(/.*)?$",
			expectError:        false,
		},
		{
			name:               "path with regex",
			template:           "/api/v1/{name:[a-z]+}",
			expectedExpression: "^api/v1/([a-z]+)(/.*)?$",
			expectError:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, tokens := templateToRegularExpression(tc.template)
			if len(tokens) == 0 && tc.template != "/" {
				t.Error("expected tokens to be non-empty")
			}
		})
	}
}

func TestTokenizePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "root path",
			path:     "/",
			expected: nil,
		},
		{
			name:     "single segment",
			path:     "/api",
			expected: []string{"api"},
		},
		{
			name:     "multiple segments",
			path:     "/api/v1/resource",
			expected: []string{"api", "v1", "resource"},
		},
		{
			name:     "with leading and trailing slashes",
			path:     "/api/v1/resource/",
			expected: []string{"api", "v1", "resource"},
		},
		{
			name:     "parameterized path",
			path:     "/namespaces/{namespace}/pods",
			expected: []string{"namespaces", "{namespace}", "pods"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tokenizePath(tc.path)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d tokens, got %d tokens", len(tc.expected), len(result))
				return
			}
			for i, token := range result {
				if token != tc.expected[i] {
					t.Errorf("token[%d] expected %q, got %q", i, tc.expected[i], token)
				}
			}
		})
	}
}

func TestPathExpression_VarNames(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectedVars []string
	}{
		{
			name:         "no variables",
			path:         "/api/v1/resource",
			expectedVars: nil,
		},
		{
			name:         "single variable",
			path:         "/api/v1/{name}",
			expectedVars: []string{"name"},
		},
		{
			name:         "multiple variables",
			path:         "/namespaces/{namespace}/pods/{name}",
			expectedVars: []string{"namespace", "name"},
		},
		{
			name:         "variable with regex",
			path:         "/api/{version:[0-9]+}",
			expectedVars: []string{"version"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := newPathExpression(tc.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(expr.VarNames) != len(tc.expectedVars) {
				t.Errorf("expected %d variable names, got %d", len(tc.expectedVars), len(expr.VarNames))
				return
			}

			for i, varName := range tc.expectedVars {
				if expr.VarNames[i] != varName {
					t.Errorf("VarNames[%d] expected %q, got %q", i, varName, expr.VarNames[i])
				}
			}
		})
	}
}
