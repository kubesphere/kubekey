//go:build builtin
// +build builtin

/*
Copyright 2025 The KubeSphere Authors.

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

package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

func TestRemoveNodesFromInventoryFile(t *testing.T) {
	tests := []struct {
		name           string
		inventory      string
		nodesToRemove  []string
		expectedOutput string
		wantErr        bool
		errMsg         string
	}{
		{
			name: "remove node from single group",
			inventory: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
        - node1
        - node2
    kube_worker:
      hosts:
        - node2
        - node3
    etcd:
      hosts:
        - node1
        - node2
`,
			nodesToRemove: []string{"node2"},
			expectedOutput: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
        - node1
    kube_worker:
      hosts:
        - node3
    etcd:
      hosts:
        - node1
`,
			wantErr: false,
		},
		{
			name: "remove multiple nodes from different groups",
			inventory: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
        - node1
        - node2
    kube_worker:
      hosts:
        - node2
        - node3
    etcd:
      hosts:
        - node1
        - node2
`,
			nodesToRemove: []string{"node2", "node3"},
			expectedOutput: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
        - node1
    kube_worker:
      hosts:
    etcd:
      hosts:
        - node1
`,
			wantErr: false,
		},
		{
			name: "remove non-existent node",
			inventory: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
  groups:
    kube_control_plane:
      hosts:
        - node1
`,
			nodesToRemove: []string{"node999"},
			wantErr:       true,
			errMsg:        "not defined in inventory",
		},
		{
			name: "remove node with 4-space indentation (no space after hosts)",
			inventory: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
      - node1
      - node2
    kube_worker:
      hosts:
      - node2
      - node3
    etcd:
      hosts:
      - node1
      - node2
`,
			nodesToRemove: []string{"node2"},
			expectedOutput: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
      - node1
    kube_worker:
      hosts:
      - node3
    etcd:
      hosts:
      - node1
`,
			wantErr: false,
		},
		{
			name: "remove node with 5-space indentation (1 space after hosts)",
			inventory: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
     - node1
     - node2
    kube_worker:
      hosts:
     - node2
     - node3
    etcd:
      hosts:
     - node1
     - node2
`,
			nodesToRemove: []string{"node2"},
			expectedOutput: `apiVersion: kubekey.kubesphere.io/v1
kind: Inventory
metadata:
  name: default
spec:
  hosts:
    node1:
    node2:
    node3:
  groups:
    kube_control_plane:
      hosts:
     - node1
    kube_worker:
      hosts:
     - node3
    etcd:
      hosts:
     - node1
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			inventoryFile := filepath.Join(tmpDir, "inventory.yaml")

			// Write initial inventory
			err := os.WriteFile(inventoryFile, []byte(tt.inventory), 0644)
			require.NoError(t, err)

			// Create DeleteNodesOptions with inventory
			o := &DeleteNodesOptions{
				CommonOptions: options.CommonOptions{
					InventoryFile: inventoryFile,
					Inventory: &kkcorev1.Inventory{
						Spec: kkcorev1.InventorySpec{
							Hosts: kkcorev1.InventoryHost{
								"node1": {Raw: []byte("{}")},
								"node2": {Raw: []byte("{}")},
								"node3": {Raw: []byte("{}")},
							},
							Groups: map[string]kkcorev1.InventoryGroup{
								"kube_control_plane": {Hosts: []string{"node1", "node2"}},
								"kube_worker":        {Hosts: []string{"node2", "node3"}},
								"etcd":               {Hosts: []string{"node1", "node2"}},
							},
						},
					},
				},
			}

			// Call the function
			err = o.removeNodesFromInventoryFile(tt.nodesToRemove)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)

			// Read the result
			content, err := os.ReadFile(inventoryFile)
			require.NoError(t, err)

			// Compare (normalize line endings)
			actual := strings.TrimSpace(string(content))
			expected := strings.TrimSpace(tt.expectedOutput)
			assert.Equal(t, expected, actual)
		})
	}
}
