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
				Host:     "local",
				Variable: &testVariable{},
			},
			exceptStderr: "unknown args for debug. only support var or msg",
		},
		{
			name: "var value",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"var": ".k"}`),
				},
				Host: "local",
				Variable: &testVariable{
					value: map[string]any{
						"k": "v",
					},
				},
			},
			exceptStdout: "v",
		},
		{
			name: "msg value",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"msg": "{{ .k }}"}`),
				},
				Host: "local",
				Variable: &testVariable{
					value: map[string]any{
						"k": "v",
					},
				},
			},
			exceptStdout: "v",
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
