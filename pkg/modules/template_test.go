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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTemplate(t *testing.T) {
	absPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		return
	}

	testcases := []struct {
		name         string
		opt          ExecOptions
		ctxFunc      func() context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "src is empty",
			opt: ExecOptions{
				Args:     runtime.RawExtension{},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctxFunc:      context.Background,
			exceptStderr: "\"src\" should be string",
		},
		{
			name: "dest is empty",
			opt: ExecOptions{
				Args: runtime.RawExtension{
					Raw: []byte(fmt.Sprintf(`{"src": "%s"}`, absPath)),
				},
				Host:     "local",
				Variable: &testVariable{},
			},
			ctxFunc:      context.Background,
			exceptStderr: "\"dest\" should be string",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctxFunc(), time.Second*5)
			defer cancel()
			acStdout, acStderr := ModuleTemplate(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, acStdout)
			assert.Equal(t, tc.exceptStderr, acStderr)
		})
	}
}
