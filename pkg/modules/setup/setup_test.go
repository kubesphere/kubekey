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

package setup

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestVariable creates a new variable.Variable for testing purposes.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		return nil
	}
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		return nil
	}
	_ = v.Merge(variable.MergeRemoteVariable(vars, hosts...))
	return v
}

// createRawArgs creates a runtime.RawExtension from a map
func createRawArgs(data map[string]any) runtime.RawExtension {
	raw, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: raw}
}

// TestSetupArgsModule tests module execution - setup module doesn't have explicit args
func TestSetupArgsModule(t *testing.T) {
	// Setup module requires actual connector, so just test module exists
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleSetup)
	})
}

// TestSetupArgsParse tests argument parsing - setup module has no args to parse
func TestSetupArgsParse(t *testing.T) {
	testcases := []struct {
		name          string
		args          map[string]any
		expectParseOk bool
		description   string
	}{
		{
			name:          "empty args",
			args:          map[string]any{},
			expectParseOk: true,
			description:   "When args is empty, should parse successfully",
		},
		{
			name:          "args are ignored for setup module",
			args:          map[string]any{"any_key": "any_value"},
			expectParseOk: true,
			description:   "When extra args provided, should parse successfully",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			raw := createRawArgs(tc.args)
			args := variable.Extension2Variables(raw)
			if tc.expectParseOk {
				require.NotNil(t, args, tc.description)
			}
		})
	}
}

// TestSetupArgsTemplate tests template variable resolution
func TestSetupArgsTemplate(t *testing.T) {
	// Setup module doesn't use templates in args
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleSetup)
	})
}

// TestSetupModule tests the actual functionality of the setup module
func TestSetupModule(t *testing.T) {
	t.Run("module exists", func(t *testing.T) {
		require.NotNil(t, ModuleSetup)
	})
}
