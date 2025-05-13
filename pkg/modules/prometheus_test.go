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
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

// Helper function to convert a map to RawExtension
func toRawExtension(obj any) runtime.RawExtension {
	bytes, err := json.Marshal(obj)
	if err != nil {
		// Handle error, log or use default value in test environment
		return runtime.RawExtension{}
	}
	return runtime.RawExtension{Raw: bytes}
}

func TestPrometheus(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		ctxFunc      func() context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "empty host",
			opt: ExecOptions{
				Args: toRawExtension(map[string]any{
					"query": "up",
				}),
				Host:     "", // Empty host
				Variable: &testVariable{},
			},
			ctxFunc:      context.Background, // Add context background
			exceptStderr: "host name is empty, please specify a prometheus host",
		},
		{
			name: "successful query",
			opt: ExecOptions{
				Args: toRawExtension(map[string]any{
					"query": "up", // Add required query parameter
				}),
				Host:     "test",
				Variable: &testVariable{},
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
			exceptStdout: "success",
		},
		{
			name: "failed query",
			opt: ExecOptions{
				Args: toRawExtension(map[string]any{
					"query": "up", // Add required query parameter
				}),
				Host:     "test",
				Variable: &testVariable{},
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, failedConnector)
			},
			exceptStderr: "failed to execute prometheus query: failed",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc

			// Safely handle ctxFunc
			if tc.ctxFunc != nil {
				ctx, cancel = context.WithTimeout(tc.ctxFunc(), time.Second*5)
				defer cancel()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
			}

			acStdout, acStderr := ModulePrometheus(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}

func TestPrometheusModuleRegistration(t *testing.T) {
	module := FindModule("prometheus")
	assert.NotNil(t, module, "Prometheus module should be registered")
}
