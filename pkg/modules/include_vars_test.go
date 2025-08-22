package modules

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestModuleIncludeVars(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
	}{
		{
			name: "include remote path",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"include_vars": "http://127.0.0.1:8080/include_vars",
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStdout: StdoutFailed,
		}, {
			name: "include empty path",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"include_vars": "",
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStdout: StdoutFailed,
		}, {
			name: "include path not exist",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"include_vars": "/path/not/exist/not_exist.yaml",
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStdout: StdoutFailed,
		}, {
			name: "include path not yaml",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"include_vars": "/path/some/exist/exist.yyy",
}`),
				},
				Variable: newTestVariable(nil, nil),
			},
			exceptStdout: StdoutFailed,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, _, _ := ModuleIncludeVars(context.Background(), testcase.opt)
			assert.Equal(t, testcase.exceptStdout, stdout)
		})
	}
}
