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
	}{
		{
			name: "successful setup with fact gathering",
			opt: ExecOptions{
				Host:     "test-host",
				Variable: newTestVariable(nil, nil),
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), ConnKey, newTestConnector(StdoutSuccess, "", nil))
			},
			exceptStdout: StdoutSuccess,
		},
		{
			name: "failed connector setup",
			opt: ExecOptions{
				Host:     "invalid-host",
				Variable: newTestVariable(nil, nil),
			},
			ctxFunc:      context.Background,
			exceptStdout: StdoutFailed,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(tc.ctxFunc(), time.Second*5)
			defer cancel()

			stdout, _, _ := ModuleSetup(ctx, tc.opt)
			assert.Equal(t, tc.exceptStdout, stdout)
		})
	}
}
