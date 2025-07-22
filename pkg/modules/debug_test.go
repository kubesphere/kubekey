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

package modules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDebug(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "non-var and non-msg",
			opt: ExecOptions{
				Args:     runtime.RawExtension{},
				Host:     "node1",
				Variable: newTestVariable(nil, nil),
			},
			exceptStderr: "\"msg\" is not found",
		},
		{
			name: "string value",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"msg": "{{ .k }}"}`),
				},
				Host: "node1",
				Variable: newTestVariable([]string{"node1"}, map[string]any{
					"k": "v",
				}),
			},
			exceptStdout: "\"v\"",
		},
		{
			name: "int value",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"msg": "{{ .k }}"}`),
				},
				Host: "node1",
				Variable: newTestVariable([]string{"node1"}, map[string]any{
					"k": 1,
				}),
			},
			exceptStdout: "1",
		},
		{
			name: "map value",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"msg": "{{ .k }}"}`),
				},
				Host: "node1",
				Variable: newTestVariable([]string{"node1"}, map[string]any{
					"k": map[string]any{
						"a1": 1,
						"a2": 1.1,
						"a3": "b3",
					},
				}),
			},
			exceptStdout: "{\n  \"a1\": 1,\n  \"a2\": 1.1,\n  \"a3\": \"b3\"\n}",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			acStdout, acStderr := ModuleDebug(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
