package modules

import (
	"context"
	"testing"
	"time"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/api/core/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestModuleAddHostvars(t *testing.T) {
	type testcase struct {
		name         string
		args         []byte
		expectStdout string
		expectStderr string
	}
	cases := []testcase{
		{
			name: "missing hosts",
			args: []byte(`
vars:
  foo: bar
`),
			expectStdout: "",
			expectStderr: "\"hosts\" should be string or string array",
		},
		{
			name: "missing vars",
			args: []byte(`
hosts: node1
`),
			expectStdout: "",
			expectStderr: "\"vars\" should not be empty",
		},
		{
			name: "invalid hosts type",
			args: []byte(`
hosts:
  foo: bar
vars:
  a: b
`),
			expectStdout: "",
			expectStderr: "\"hosts\" should be string or string array",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			opt := ExecOptions{
				Args:     runtime.RawExtension{Raw: tc.args},
				Host:     "",
				Variable: &testVariable{},
				Task:     kkcorev1alpha1.Task{},
				Playbook: kkcorev1.Playbook{},
			}
			stdout, stderr := ModuleAddHostvars(ctx, opt)
			require.Equal(t, tc.expectStdout, stdout, "stdout mismatch")
			require.Equal(t, tc.expectStderr, stderr, "stderr mismatch")
		})
	}
}
