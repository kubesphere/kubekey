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
					"localhost": {
						RemoteVars: map[string]any{},
						RuntimeVars: map[string]any{
							"k1": "k2",
						},
					},
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

func TestEnsureLocalHostSyncsRuntimeVars(t *testing.T) {
	v := &variable{
		source: source.NewMemorySource(),
		value: &value{
			Hosts: map[string]host{
				"test": {
					RuntimeVars: map[string]any{
						"certs": map[string]any{
							"ca": map[string]any{
								"date":            "87600h",
								"gen_cert_policy": "IfNotPresent",
							},
						},
						"binary_dir": "/data/kubekey",
					},
				},
			},
		},
	}
	if err := v.Merge(EnsureLocalHost()); err != nil {
		t.Fatal(err)
	}
	got, err := v.Get(GetAllVariable("localhost"))
	if err != nil {
		t.Fatal(err)
	}
	vars, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got %T, want map[string]any", got)
	}
	certs, ok := vars["certs"].(map[string]any)
	if !ok {
		t.Fatalf("localhost missing certs, got %#v", vars)
	}
	ca, ok := certs["ca"].(map[string]any)
	if !ok || ca["gen_cert_policy"] != "IfNotPresent" {
		t.Fatalf("unexpected certs.ca: %#v", certs["ca"])
	}
	if vars["binary_dir"] != "/data/kubekey" {
		t.Fatalf("binary_dir = %v, want /data/kubekey", vars["binary_dir"])
	}
}

func TestMergeRuntimeVariableSyncsLocalhost(t *testing.T) {
	node, err := converter.ConvertMap2Node(map[string]any{
		"certs": map[string]any{
			"ca": map[string]any{
				"gen_cert_policy": "IfNotPresent",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	v := &variable{
		source: source.NewMemorySource(),
		value: &value{
			Hosts: map[string]host{
				"test": {},
			},
		},
	}
	if err := v.Merge(MergeRuntimeVariable([]yaml.Node{node}, "test")); err != nil {
		t.Fatal(err)
	}

	got, err := v.Get(GetAllVariable("localhost"))
	if err != nil {
		t.Fatal(err)
	}
	vars, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got %T, want map[string]any", got)
	}
	certs, ok := vars["certs"].(map[string]any)
	if !ok {
		t.Fatalf("localhost missing certs, got %#v", vars)
	}
	ca, ok := certs["ca"].(map[string]any)
	if !ok || ca["gen_cert_policy"] != "IfNotPresent" {
		t.Fatalf("unexpected certs.ca: %#v", certs["ca"])
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
