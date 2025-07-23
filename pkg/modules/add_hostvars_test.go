package modules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestModuleAddHostvars(t *testing.T) {
	type testcase struct {
		name         string
		opt          ExecOptions
		expectStdout string
		expectStderr string
	}
	cases := []testcase{
		{
			name: "missing hosts",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
vars:
  foo: bar
`),
				},
			},
			expectStdout: "",
			expectStderr: "\"hosts\" should be string or string array",
		},
		{
			name: "missing vars",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
`),
				},
			},
			expectStdout: "",
			expectStderr: "\"vars\" should not be empty",
		},
		{
			name: "invalid hosts type",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts:
  foo: bar
vars:
  a: b
`),
				},
			},
			expectStdout: "",
			expectStderr: "\"hosts\" should be string or string array",
		},
		{
			name: "string value",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  a: b
`),
				},
			},
		},
		{
			name: "string var value",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, map[string]any{"a": "b"}),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  a: "{{ .a }}"		
`),
				},
			},
		},
		{
			name: "map value",
			opt: ExecOptions{
				Host:     "node1",
				Variable: newTestVariable([]string{"node1"}, nil),
				Args: runtime.RawExtension{
					Raw: []byte(`
hosts: node1
vars:
  a:
    b: c
`),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, stderr := ModuleAddHostvars(ctx, tc.opt)
			require.Equal(t, tc.expectStdout, stdout, "stdout mismatch")
			require.Equal(t, tc.expectStderr, stderr, "stderr mismatch")
		})
	}
}
