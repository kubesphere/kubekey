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
			exceptStdout: StdoutFailed,
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
			exceptStdout: StdoutFailed,
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
			exceptStdout: StdoutFailed,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, _, _ := ModuleImage(context.Background(), testcase.opt)
			assert.Equal(t, testcase.exceptStdout, stdout)
		})
	}
}
