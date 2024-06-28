/*
Copyright 2024 The KubeSphere Authors.

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

func TestFetch(t *testing.T) {
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
				Variable: &testVariable{},
			},
			ctx: context.WithValue(context.Background(), "connector", &testConnector{
				output: []byte("success"),
			}), exceptStderr: "\"src\" in args should be string",
		},
		{
			name: "dest is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"src": "/etc/test.txt"}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctx: context.WithValue(context.Background(), "connector", &testConnector{
				output: []byte("success"),
			}),
			exceptStderr: "\"dest\" in args should be string",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctx, time.Second*5)
			defer cancel()
			acStdout, acStderr := ModuleFetch(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
