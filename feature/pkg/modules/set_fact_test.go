/*
Copyright 2023 The KubeSphere Authors.

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	kkcorev1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1"
	kkcorev1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/core/v1alpha1"
)

func TestSetFact(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "success",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(`{"k": "v"}`),
				},
				Host:     "",
				Variable: &testVariable{},
				Task:     kkcorev1alpha1.Task{},
				Pipeline: kkcorev1.Pipeline{},
			},
			exceptStdout: "success",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			stdout, stderr := ModuleSetFact(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, stdout)
			assert.Equal(t, tc.exceptStderr, stderr)
		})
	}
}
