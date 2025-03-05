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

func TestCommand(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		ctxFunc      func() context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "non-host variable",
			opt: ExecOptions{
				Variable: &testVariable{},
			},
			ctxFunc:      context.Background,
			exceptStderr: "failed to connector of \"\" error: host is not set",
		},
		{
			name: "exec command success",
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
			opt: ExecOptions{
				Host:     "test",
				Args:     runtime.RawExtension{Raw: []byte("echo success")},
				Variable: &testVariable{},
			},
			exceptStdout: "success",
		},
		{
			name:    "exec command failed",
			ctxFunc: func() context.Context { return context.WithValue(context.Background(), ConnKey, failedConnector) },
			opt: ExecOptions{
				Host:     "test",
				Args:     runtime.RawExtension{Raw: []byte("echo success")},
				Variable: &testVariable{},
			},
			exceptStderr: "failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctxFunc(), time.Second*5)
			defer cancel()

			acStdout, acStderr := ModuleCommand(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
