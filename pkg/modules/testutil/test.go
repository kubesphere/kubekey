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

// Package testutil provides helpers shared by pkg/modules/* unit and
// integration tests. They let a test build a module execution context
// (variable.Variable) without a real cluster.
package testutil

import (
	"context"

	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

// NewTestVariable creates a new variable.Variable for testing purposes.
// It initializes a test playbook and client via _const.NewTestPlaybook, then
// creates an in-memory variable source. It merges the provided vars as remote variables
// for the specified hosts. This allows you to customize the per-host variable context
// for unit tests needing module execution context.
func NewTestVariable(hosts []string, vars map[string]any) variable.Variable {
	client, playbook, err := _const.NewTestPlaybook(hosts)
	if err != nil {
		// If creating the test playbook failed, log the error and continue (returned Variable may be nil)
		klog.ErrorS(err, "failed to create test playbook")
	}
	// Create a new variable in memory using the test client and playbook
	v, err := variable.New(context.TODO(), client, *playbook, source.MemorySource)
	if err != nil {
		// If creating the variable failed, log and return what we have (likely nil)
		klog.ErrorS(err, "failed to create variable")
	}
	// Set default values by merging the provided vars as remote variables for the hosts.
	if err := v.Merge(variable.MergeRemoteVariable(vars, hosts...)); err != nil {
		// If merging variables failed, log the error.
		klog.ErrorS(err, "failed to merge variable")
	}

	return v
}
