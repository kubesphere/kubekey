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

func TestAssert(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "non-that",
			opt: ExecOptions{
				Host:     "local",
				Variable: &testVariable{},
				Args:     runtime.RawExtension{},
			},
			exceptStderr: "\"that\" should be []string or string",
		},
		{
			name: "success with non-msg",
			opt: ExecOptions{
				Host: "local",
				Args: runtime.RawExtension{
					Raw: []byte(`{"that": ["true", "eq .testvalue \"a\""]}`),
				},
				Variable: &testVariable{
					value: map[string]any{
						"testvalue": "a",
					},
				},
			},
			exceptStdout: StdoutTrue,
		},
		{
			name: "success with success_msg",
			opt: ExecOptions{
				Host: "local",
				Args: runtime.RawExtension{
					Raw: []byte(`{"that": ["true", "eq .k1 \"v1\""], "success_msg": "success {{ .k2 }}"}`),
				},
				Variable: &testVariable{
					value: map[string]any{
						"k1": "v1",
						"k2": "v2",
					},
				},
			},
			exceptStdout: "success v2",
		},
		{
			name: "failed with non-msg",
			opt: ExecOptions{
				Host: "local",
				Args: runtime.RawExtension{
					Raw: []byte(`{"that": ["true", "eq .k1 \"v2\""]}`),
				},
				Variable: &testVariable{
					value: map[string]any{
						"k1": "v1",
						"k2": "v2",
					},
				},
			},
			exceptStdout: StdoutFalse,
			exceptStderr: "False",
		},
		{
			name: "failed with failed_msg",
			opt: ExecOptions{
				Host: "local",
				Args: runtime.RawExtension{
					Raw: []byte(`{"that": ["true", "eq .k1 \"v2\""], "fail_msg": "failed {{ .k2 }}"}`),
				},
				Variable: &testVariable{
					value: map[string]any{
						"k1": "v1",
						"k2": "v2",
					},
				},
			},
			exceptStdout: StdoutFalse,
			exceptStderr: "failed v2",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			acStdout, acStderr := ModuleAssert(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
