package modules

import (
	"context"
	"os"
	"testing"

	testassert "github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestModuleGenCert(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "gen root cert",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{
"policy": "IfNotPresent",
"sans": ["localhost"],
"cn": "test", 
"out_key": "./test_gen_cert/test-key.pem",
"out_cert": "./test_gen_cert/test-crt.pem"
				}`),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			exceptStdout: "success",
		},
	}

	if _, err := os.Stat("./test_gen_cert"); os.IsNotExist(err) {
		if err := os.Mkdir("./test_gen_cert", 0755); err != nil {
			t.Fatal(err)
		}
	}
	defer os.RemoveAll("./test_gen_cert")

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, stderr := ModuleGenCert(context.Background(), testcase.opt)
			testassert.Equal(t, testcase.exceptStdout, stdout)
			testassert.Equal(t, testcase.exceptStderr, stderr)
		})
	}
}
