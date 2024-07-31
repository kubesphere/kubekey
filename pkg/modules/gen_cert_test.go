/*
Copyright 2024 The KubeSphere Authors.

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

package modules

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
					Raw: []byte(`
policy: |
  {{- .policy -}}
sans: ["localhost"]
cn: "test"
out_key: ./test_gen_cert/test-key.pem
out_cert: ./test_gen_cert/test-crt.pem
`),
				},
				Host: "local",
				Variable: &testVariable{
					value: map[string]any{
						"policy": "IfNotPresent",
					},
				},
			},
			exceptStdout: "success",
		},
	}

	if _, err := os.Stat("./test_gen_cert"); os.IsNotExist(err) {
		if err := os.Mkdir("./test_gen_cert", os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	defer os.RemoveAll("./test_gen_cert")

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			stdout, stderr := ModuleGenCert(context.Background(), testcase.opt)
			assert.Equal(t, testcase.exceptStdout, stdout)
			assert.Equal(t, testcase.exceptStderr, stderr)
		})
	}
}
