package modules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestModuleImage(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "pull is not map",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"pull": ""
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStderr: "\"pull\" should be map",
		},
		{
			name: "pull.manifests is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"pull": {}
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStderr: "\"pull.manifests\" is required",
		},
		{
			name: "push is not map",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"push": ""
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStderr: "\"push\" should be map",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, stderr := ModuleImage(context.Background(), testcase.opt)
			assert.Equal(t, testcase.exceptStdout, stdout)
			assert.Equal(t, testcase.exceptStderr, stderr)
		})
	}
}
