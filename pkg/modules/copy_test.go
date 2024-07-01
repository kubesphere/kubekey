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

func TestCopy(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		ctxFunc      func() context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "src and content is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"dest": "hello world"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
			exceptStderr: "either \"src\" or \"content\" must be provided.",
		},
		{
			name: "dest is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"content": "hello world"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
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
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
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
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
			exceptStdout: stdoutSuccess,
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
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, failedConnector)
			},
			exceptStderr: "copy file error: failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctxFunc(), time.Second*5)
			defer cancel()
			acStdout, acStderr := ModuleCopy(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
