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

package connector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/exec"
)

func TestLocalConnector_ExecuteCommand(t *testing.T) {
	testcases := []struct {
		name           string
		cmd            string
		localConnector *localConnector
		expectedStdout string
	}{
		{
			name:           "execute command succeed",
			cmd:            "echo 'hello'",
			localConnector: &localConnector{Cmd: exec.New()},
			expectedStdout: "hello\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			a, _ := tc.localConnector.ExecuteCommand(ctx, tc.cmd)
			assert.Equal(t, tc.expectedStdout, string(a))
		})
	}
}
