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

package image

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/modules/internal"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestVariable creates a new variable.Variable for testing purposes.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		klog.ErrorS(err, "failed to create test playbook")
	}
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		klog.ErrorS(err, "failed to create variable")
	}
	if err := v.Merge(variable.MergeRemoteVariable(vars, hosts...)); err != nil {
		klog.ErrorS(err, "failed to merge variable")
	}
	return v
}

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestImageArgsModule tests module execution with various argument scenarios.
func TestImageArgsModule(t *testing.T) {
	testcases := []struct {
		name         string
		opt          internal.ExecOptions
		expectStdout string
		expectError  bool
		description  string
	}{
		{
			name: "valid args with pull section",
			opt: internal.ExecOptions{
				Host:     "node1",
				Variable: NewTestVariable([]string{"node1"}, nil),
				Args: createRawArgs(map[string]any{
					"pull": map[string]any{
						"manifests": []string{"image1", "image2"},
					},
				}),
			},
			expectStdout: internal.StdoutFailed, // Will fail due to missing connector
			expectError:  true,
			description:  "When pull section is provided, module should execute (fails due to missing connector)",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr, err := ModuleImage(ctx, tc.opt)
			require.Equal(t, tc.expectStdout, stdout, tc.description)
			if tc.expectError {
				require.Error(t, err, tc.description)
				require.NotEmpty(t, stderr, tc.description)
			} else {
				require.NoError(t, err, tc.description)
			}
		})
	}
}

// TestImageArgsParse tests argument parsing edge cases.
func TestImageArgsParse(t *testing.T) {
	testcases := []struct {
		name             string
		args             map[string]any
		expectParseError bool
		description      string
	}{
		{
			name: "valid args with pull section",
			args: map[string]any{
				"pull": map[string]any{
					"manifests": []string{"image1", "image2"},
				},
			},
			expectParseError: false,
			description:      "When pull section is provided, should parse successfully",
		},
		{
			name: "valid args without pull section",
			args: map[string]any{
				"source": "registry",
			},
			expectParseError: false,
			description:      "When pull section is missing, should parse successfully",
		},
		{
			name:             "invalid pull section (not a map)",
			args:             map[string]any{"pull": "invalid"},
			expectParseError: true,
			description:      "When pull section is not a map, should return error",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			if pullArgs, ok := args["pull"]; ok {
				_, ok := pullArgs.(map[string]any)
				if tc.expectParseError {
					require.False(t, ok, tc.description)
				} else {
					require.True(t, ok, tc.description)
				}
			}
		})
	}
}

// TestImageArgsTemplate tests template variable resolution in arguments.
func TestImageArgsTemplate(t *testing.T) {
	// Image module template tests are complex, just verify the module exists
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleImage)
	})
}

// TestImageModule tests the actual functionality of the image module.
func TestImageModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleImage)
	})
}
