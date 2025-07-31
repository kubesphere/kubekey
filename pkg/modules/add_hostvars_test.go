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
			expectStdout: StdoutFailed,
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
			expectStdout: StdoutFailed,
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
			expectStdout: StdoutFailed,
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
			expectStdout: StdoutSuccess,
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
			expectStdout: StdoutSuccess,
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
			expectStdout: StdoutSuccess,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			stdout, _, _ := ModuleAddHostvars(ctx, tc.opt)
			require.Equal(t, tc.expectStdout, stdout, "stdout mismatch")
		})
	}
}
