package variable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/kubesphere/kubekey/v4/pkg/converter"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

func TestMergeRemoteVariable(t *testing.T) {
	testcases := []struct {
		name     string
		variable *variable
		hostname string
		data     map[string]any
		except   value
	}{
		{
			name: "success",
			variable: &variable{
				source: source.NewMemorySource(),
				value: &value{
					Hosts: map[string]host{
						"n1": {},
						"n2": {},
					},
				},
			},
			hostname: "n1",
			data: map[string]any{
				"k1": "k2",
			},
			except: value{
				Hosts: map[string]host{
					"n1": {
						RemoteVars: map[string]any{
							"k1": "k2",
						},
					},
					"n2": {},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.variable.Merge(MergeRemoteVariable(tc.data, tc.hostname))
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, *tc.variable.value)
		})
	}
}

func TestMergeRuntimeVariable(t *testing.T) {
	testcases := []struct {
		name     string
		variable *variable
		hostname string
		data     map[string]any
		except   value
	}{
		{
			name: "success",
			variable: &variable{
				source: source.NewMemorySource(),
				value: &value{
					Hosts: map[string]host{
						"n1": {},
						"n2": {},
					},
				},
			},
			hostname: "n1",
			data: map[string]any{
				"k1": "k2",
			},
			except: value{
				Hosts: map[string]host{
					"n1": {
						RuntimeVars: map[string]any{
							"k1": "k2",
						},
					},
					"n2": {},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := converter.ConvertMap2Node(tc.data)
			if err != nil {
				t.Fatal(err)
			}
			if err := tc.variable.Merge(MergeRuntimeVariable([]yaml.Node{node}, tc.hostname)); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, *tc.variable.value)
		})
	}
}

func TestMergeResultVariable(t *testing.T) {
	testcases := []struct {
		name     string
		host     string
		variable *variable
		data     map[string]any
		except   value
	}{
		{
			name: "success",
			host: "n1",
			variable: &variable{
				source: source.NewMemorySource(),
				value: &value{
					Hosts: map[string]host{
						"n1": {
							RuntimeVars: map[string]any{
								"k1": "v1",
							},
						},
						"n2": {
							RuntimeVars: map[string]any{
								"k1": "v2",
							},
						},
					},
				},
			},
			data: map[string]any{
				"v1": "v1",
				"v2": "vv",
			},
			except: value{
				Hosts: map[string]host{
					"n1": {
						RuntimeVars: map[string]any{
							"k1": "v1",
						},
					},
					"n2": {
						RuntimeVars: map[string]any{
							"k1": "v2",
						},
					},
				},
				Result: map[string]any{
					resultKey: map[string]any{
						"v1": "v1",
						"v2": "vv",
					},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.variable.Merge(MergeResultVariable(tc.data)); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.except, *tc.variable.value)
		})
	}
}
