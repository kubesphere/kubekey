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

	"github.com/cockroachdb/errors"
	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
)

/*
The Assert module evaluates boolean conditions and returns success/failure messages.
This module allows users to perform assertions in playbooks and control flow based on conditions.

Configuration:
Users can specify conditions to evaluate and customize success/failure messages:

assert:
  that:                    # List of conditions to evaluate (required)
    - "{{ condition1 }}"
    - "{{ condition2 }}"
  success_msg: "Success"   # optional: message to return on success (default: "True")
  fail_msg: "Failed"       # optional: high priority failure message
  msg: "Error"            # optional: fallback failure message (default: "False")

Usage Examples in Playbook Tasks:
1. Basic assertion:
   ```yaml
   - name: Check if service is running
     assert:
       that:
         - "{{ service_status == 'running' }}"
     register: check_result
   ```

2. Custom messages:
   ```yaml
   - name: Verify deployment
     assert:
       that:
         - "{{ deployment_ready }}"
         - "{{ pods_running }}"
       success_msg: "Deployment is healthy"
       fail_msg: "Deployment check failed"
     register: verify_result
   ```

3. Multiple conditions:
   ```yaml
   - name: Validate configuration
     assert:
       that:
         - "{{ config_valid }}"
         - "{{ required_fields_present }}"
         - "{{ values_in_range }}"
       msg: "Configuration validation failed"
     register: config_check
   ```

Return Values:
- On success: Returns success_msg (or "True") in stdout
- On failure: Returns fail_msg (or msg or "False") in stderr
*/

type assertArgs struct {
	that       []string
	successMsg string
	failMsg    string // high priority than msg
	msg        string
}

func newAssertArgs(_ context.Context, raw runtime.RawExtension, vars map[string]any) (*assertArgs, error) {
	var err error
	var aa = &assertArgs{}

	args := variable.Extension2Variables(raw)
	if aa.that, err = variable.StringSliceVar(vars, args, "that"); err != nil {
		return nil, errors.New("\"that\" should be []string or string")
	}
	for i, s := range aa.that {
		if !kkprojectv1.IsTmplSyntax(s) {
			aa.that[i] = kkprojectv1.ParseTmplSyntax(s)
		}
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

// ModuleAssert handles the "assert" module, evaluating boolean conditions and returning appropriate messages
func ModuleAssert(ctx context.Context, options ExecOptions) (string, string) {
	// get host variable
	ha, err := options.getAllVariables()
	if err != nil {
		return "", err.Error()
	}

	aa, err := newAssertArgs(ctx, options.Args, ha)
	if err != nil {
		return "", err.Error()
	}

	ok, err := tmpl.ParseBool(ha, aa.that...)
	if err != nil {
		return "", fmt.Sprintf("parse \"that\" error: %v", err)
	}
	// condition is true
	if ok {
		r, err := tmpl.Parse(ha, aa.successMsg)
		if err == nil {
			return string(r), ""
		}
		klog.V(4).ErrorS(err, "parse \"success_msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))

		return StdoutTrue, ""
	}
	// condition is false and fail_msg is not empty
	if aa.failMsg != "" {
		r, err := tmpl.Parse(ha, aa.failMsg)
		if err == nil {
			return StdoutFalse, string(r)
		}
		klog.V(4).ErrorS(err, "parse \"fail_msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))
	}
	// condition is false and msg is not empty
	if aa.msg != "" {
		r, err := tmpl.Parse(ha, aa.msg)
		if err == nil {
			return StdoutFalse, string(r)
		}
		klog.V(4).ErrorS(err, "parse \"msg\" error", "task", ctrlclient.ObjectKeyFromObject(&options.Task))
	}

	return StdoutFalse, "False"
}

func init() {
	utilruntime.Must(RegisterModule("assert", ModuleAssert))
}
