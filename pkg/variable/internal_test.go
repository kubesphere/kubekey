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

func TestMergeGroup(t *testing.T) {
	testcases := []struct {
		name   string
		g1     []string
		g2     []string
		except []string
	}{
		{
			name: "non-repeat",
			g1: []string{
				"h1", "h2", "h3",
			},
			g2: []string{
				"h4", "h5",
			},
			except: []string{
				"h1", "h2", "h3", "h4", "h5",
			},
		},
		{
			name: "repeat value",
			g1: []string{
				"h1", "h2", "h3",
			},
			g2: []string{
				"h3", "h4", "h5",
			},
			except: []string{
				"h1", "h2", "h3", "h4", "h5",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ac := mergeSlice(tc.g1, tc.g2)
			assert.Equal(t, tc.except, ac)
		})
	}
}
