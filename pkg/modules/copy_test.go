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
	"fmt"
	"testing"
	"time"

	testassert "github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCopy(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		ctx          context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "src and content is empty",
			opt: ExecOptions{
				Args:     runtime.RawExtension{},
				Host:     "local",
				Variable: nil,
			},
			ctx:          context.Background(),
			exceptStderr: "\"src\" or \"content\" in args should be string",
		},
		{
			name: "dest is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"content": "hello world"}`),
				},
				Host:     "local",
				Variable: nil,
			},
			ctx:          context.Background(),
			exceptStderr: "\"dest\" in args should be string",
		},
		{
			name: "content not copy to file",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"content": "hello world", "dest": "/etc/"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctx: context.WithValue(context.Background(), "connector", &testConnector{
				output: []byte("success"),
			}),
			exceptStderr: "\"content\" should copy to a file",
		},
		{
			name: "copy success",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"content": "hello world", "dest": "/etc/test.txt"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctx: context.WithValue(context.Background(), "connector", &testConnector{
				output: []byte("success"),
			}),
			exceptStdout: "success",
		},
		{
			name: "copy failed",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"content": "hello world", "dest": "/etc/test.txt"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctx: context.WithValue(context.Background(), "connector", &testConnector{
				copyErr: fmt.Errorf("copy failed"),
			}),
			exceptStderr: "copy failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctx, time.Second*5)
			defer cancel()
			acStdout, acStderr := ModuleCopy(ctx, tc.opt)
			testassert.Equal(t, tc.exceptStdout, acStdout)
			testassert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
