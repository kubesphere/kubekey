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

package converter

import (
	"testing"

	capkkinfrav1beta1 "github.com/kubesphere/kubekey/api/capkk/infrastructure/v1beta1"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGroupHostBySerial(t *testing.T) {
	hosts := []string{"h1", "h2", "h3", "h4", "h5", "h6", "h7"}
	testcases := []struct {
		name         string
		serial       []any
		exceptResult [][]string
		exceptErr    bool
	}{
		{
			name:   "group host by 1",
			serial: []any{1},
			exceptResult: [][]string{
				{"h1"},
				{"h2"},
				{"h3"},
				{"h4"},
				{"h5"},
				{"h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 2",
			serial: []any{2},
			exceptResult: [][]string{
				{"h1", "h2"},
				{"h3", "h4"},
				{"h5", "h6"},
				{"h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  2",
			serial: []any{1, 2},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3"},
				{"h4", "h5"},
				{"h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:   "group host by serial 1 and  40%",
			serial: []any{"1", "40%"},
			exceptResult: [][]string{
				{"h1"},
				{"h2", "h3", "h4"},
				{"h5", "h6", "h7"},
			},
			exceptErr: false,
		},
		{
			name:         "group host by unSupport serial type",
			serial:       []any{1.1},
			exceptResult: nil,
			exceptErr:    true,
		},
		{
			name:         "group host by unSupport serial value",
			serial:       []any{"%1.1%"},
			exceptResult: nil,
			exceptErr:    true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GroupHostBySerial(hosts, tc.serial)
			if err != nil {
				if tc.exceptErr {
					assert.Error(t, err)

					return
				}
				t.Fatal(err)
			}
			assert.Equal(t, tc.exceptResult, result)
		})
	}
}

func TestConvertKKClusterToInventoryHost(t *testing.T) {
	testcases := []struct {
		name      string
		kkcluster *capkkinfrav1beta1.KKCluster
		except    kkcorev1.InventoryHost
	}{
		{

			name: "test success",
			kkcluster: &capkkinfrav1beta1.KKCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: capkkinfrav1beta1.KKClusterSpec{
					InventoryHosts: []capkkinfrav1beta1.InventoryHost{
						{
							Name: "h1",
							Connector: capkkinfrav1beta1.InventoryHostConnector{
								Type: "local",
								Host: "127.0.0.1",
							},
							Vars: runtime.RawExtension{
								Raw: []byte(`{"internal_ipv4":"127.0.1.1"}`),
							},
						},
						{
							Name: "h2",
							Connector: capkkinfrav1beta1.InventoryHostConnector{
								Type: "ssh",
								Host: "127.0.0.2",
							},
							Vars: runtime.RawExtension{
								Raw: []byte(`{"internal_ipv4":"127.0.1.2"}`),
							},
						},
					},
				},
			},
			except: kkcorev1.InventoryHost{
				"h1": runtime.RawExtension{
					Raw: []byte(`{"connector":{"type":"local","host":"127.0.0.1"},"internal_ipv4":"127.0.1.1"}`),
				},
				"h2": runtime.RawExtension{
					Raw: []byte(`{"connector":{"type":"ssh","host":"127.0.0.2"},"internal_ipv4":"127.0.1.2"}`),
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ConvertKKClusterToInventoryHost(tc.kkcluster)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.except, result)
		})
	}
}

func TestConvertMap2Node(t *testing.T) {
	testcases := []struct {
		name    string
		input   map[string]any
		wantErr bool
	}{
		{
			name: "simple map",
			input: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
		},
		{
			name: "nested map",
			input: map[string]any{
				"outer": map[string]any{
					"inner": "value",
				},
				"array": []any{"a", "b", "c"},
			},
		},
		{
			name:  "empty map",
			input: map[string]any{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := ConvertMap2Node(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			assert.NotNil(t, node)

			// Convert back to map to verify roundtrip
			var result map[string]any
			err = node.Decode(&result)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.input, result)
		})
	}
}
