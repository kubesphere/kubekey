package modules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModuleSetup(t *testing.T) {
	testcases := []struct {
		name         string
		opt          ExecOptions
		ctxFunc      func() context.Context
		exceptStdout string
		exceptStderr string
	}{
		{
			name: "successful setup with fact gathering",
			opt: ExecOptions{
				Host:     "test-host",
				Variable: &testVariable{},
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, successConnector)
			},
			exceptStdout: StdoutSuccess,
			exceptStderr: "",
		},
		{
			name: "failed connector setup",
			opt: ExecOptions{
				Host:     "invalid-host",
				Variable: &testVariable{},
			},
			ctxFunc:      context.Background,
			exceptStdout: "",
			exceptStderr: "failed to connector of \"invalid-host\" error:",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctxFunc(), time.Second*5)
			defer cancel()

			stdout, stderr := ModuleSetup(ctx, tc.opt)
			assert.Contains(t, stdout, tc.exceptStdout)
			if tc.exceptStderr != "" {
				assert.Contains(t, stderr, tc.exceptStderr)
			} else {
				assert.Empty(t, stderr)
			}
		})
	}
}
