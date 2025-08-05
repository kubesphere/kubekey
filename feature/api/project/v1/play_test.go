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

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalSerial(t *testing.T) {
	testcases := []struct {
		name    string
		content string
		except  []any
	}{
		{
			name: "test single string",
			content: `
host1`,
			except: []any{
				"host1",
			},
		},
		{
			name: "test single number",
			content: `
1`,
			except: []any{
				"1",
			},
		},
		{
			name: "test single percent",
			content: `
10%`,
			except: []any{
				"10%",
			},
		},
		{
			name: "test multi value",
			content: ` 
- host1
- 1
- 10%`,
			except: []any{
				"host1",
				1,
				"10%",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var serial PlaySerial
			err := yaml.Unmarshal([]byte(tc.content), &serial)
			assert.NoError(t, err)
			assert.Equal(t, tc.except, serial.Data)
		})
	}
}
