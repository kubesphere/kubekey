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

func TestToBool(t *testing.T) {
	testcases := []struct {
		name     string
		input    any
		expected bool
		wantErr  bool
	}{
		{
			name:     "bool true",
			input:    true,
			expected: true,
		},
		{
			name:     "bool false",
			input:    false,
			expected: false,
		},
		{
			name:     "string true",
			input:    "true",
			expected: true,
		},
		{
			name:     "string false",
			input:    "false",
			expected: false,
		},
		{
			name:     "string 1",
			input:    "1",
			expected: true,
		},
		{
			name:     "string 0",
			input:    "0",
			expected: false,
		},
		{
			name:     "int non-zero",
			input:    42,
			expected: true,
		},
		{
			name:     "int zero",
			input:    0,
			expected: false,
		},
		{
			name:    "invalid string",
			input:   "yes",
			wantErr: true,
		},
		{
			name:    "nil",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   map[string]string{},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toBool(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
