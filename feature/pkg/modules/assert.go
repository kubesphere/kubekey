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
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

type assertArgs struct {
	that       []string
	successMsg string
	failMsg    string // high priority than msg
	msg        string
}

func newAssertArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*assertArgs, error) {
	var err error
	aa := &assertArgs{}
	args := variable.Extension2Variables(raw)
	if aa.that, err = variable.StringSliceVar(vars, args, "that"); err != nil {
		return nil, errors.New("\"that\" should be []string or string")
	}
	aa.successMsg, _ = variable.StringVar(vars, args, "success_msg")
	if aa.successMsg == "" {
		aa.successMsg = StdoutTrue
	}
	aa.failMsg, _ = variable.StringVar(vars, args, "fail_msg")
	aa.msg, _ = variable.StringVar(vars, args, "msg")
	if aa.msg == "" {
		aa.msg = StdoutFalse
	}

	return aa, nil
}

// ModuleAssert deal "assert" module
func ModuleAssert(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	aa, err := newAssertArgs(ctx, options.Args, ha)
	if err != nil {
		klog.V(4).ErrorS(err, "get assert args error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))

		return "", err.Error()
	}

	ok, err := tmpl.ParseBool(ha, aa.that)
	if err != nil {
		return "", fmt.Sprintf("parse \"that\" error: %v", err)
	}
	// condition is true
	if ok {
		r, err := tmpl.ParseString(ha, aa.successMsg)
		if err == nil {
			return r, ""
		}
		klog.V(4).ErrorS(err, "parse \"success_msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))

		return StdoutTrue, ""
	}
	// condition is false and fail_msg is not empty
	if aa.failMsg != "" {
		r, err := tmpl.ParseString(ha, aa.failMsg)
		if err == nil {
			return StdoutFalse, r
		}
		klog.V(4).ErrorS(err, "parse \"fail_msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))
	}
	// condition is false and msg is not empty
	if aa.msg != "" {
		r, err := tmpl.ParseString(ha, aa.msg)
		if err == nil {
			return StdoutFalse, r
		}
		klog.V(4).ErrorS(err, "parse \"msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))
	}

	return StdoutFalse, "False"
}
