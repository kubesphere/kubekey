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

package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupHostBySerial(t *testing.T) {
	hosts := []string{"h1", "h2", "h3", "h4", "h5", "h6", "h7"}
	testcases := []struct {
		name         string
		serial       []any
		exceptResult [][]string
		exceptErr    bool
	}{
		{
			name:   "group host by 1",
			serial: []any{1},
			exceptResult: [][]string{
				{"h1"},
				{"h2"},
				{"h3"},
				{"h4"},
				{"h5"},
				{"h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 2",
			serial: []any{2},
			exceptResult: [][]string{
				{"h1", "h2"},
				{"h3", "h4"},
				{"h5", "h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  2",
			serial: []any{1, 2},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3"},
				{"h4", "h5"},
				{"h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  40%",
			serial: []any{"1", "40%"},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3", "h4"},
				{"h5", "h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:         "group host by unSupport serial type",
			serial:       []any{1.1},
			exceptResult: nil,
			exceptErr:    true,
		},
		{
			name:         "group host by unSupport serial value",
			serial:       []any{"%1.1%"},
			exceptResult: nil,
			exceptErr:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GroupHostBySerial(hosts, tc.serial)
			if err != nil {
				if tc.exceptErr {
					assert.Error(t, err)

					return
				}
				t.Fatal(err)
			}
			assert.Equal(t, tc.exceptResult, result)
		})
	}
}
