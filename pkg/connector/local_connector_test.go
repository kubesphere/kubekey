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
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/exec"
	testingexec "k8s.io/utils/exec/testing"
)

func newFakeLocalConnector(runCmd string, output string) *localConnector {
	return &localConnector{
		Cmd: &testingexec.FakeExec{CommandScript: []testingexec.FakeCommandAction{
			func(cmd string, args ...string) exec.Cmd {
				if strings.TrimSpace(fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))) == "/bin/sh -c "+runCmd {
					return &testingexec.FakeCmd{
						CombinedOutputScript: []testingexec.FakeAction{func() ([]byte, []byte, error) {
							return []byte(output), nil, nil
						}},
					}
				}

				return &testingexec.FakeCmd{
					CombinedOutputScript: []testingexec.FakeAction{func() ([]byte, []byte, error) {
						return nil, nil, errors.New("error command")
					}},
				}
			},
		}},
	}
}

func TestSshConnector_ExecuteCommand(t *testing.T) {
	testcases := []struct {
		name        string
		cmd         string
		exceptedErr error
	}{
		{
			name:        "execute command succeed",
			cmd:         "echo 'hello'",
			exceptedErr: nil,
		},
		{
			name:        "execute command failed",
			cmd:         "echo 'hello1'",
			exceptedErr: errors.New("error command"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			lc := newFakeLocalConnector("echo 'hello'", "hello")
			_, err := lc.ExecuteCommand(ctx, tc.cmd)
			assert.Equal(t, tc.exceptedErr, err)
		})
	}
}
