//go:build builtin
// +build builtin

/*
Copyright 2026 The KubeSphere Authors.

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

// Package core_test. The dual-stack work made precheck reject configurations it
// used to accept, so this test replays the real
// builtin/core/roles/precheck/network/tasks/main.yaml — every "when" and every
// "that", in file order, evaluated through the same path the assert module uses
// (kkprojectv1.ParseTmplSyntax + tmpl.ParseBool) — against a set of inventories.
//
// The conditions are read from the playbook rather than copied, so editing the
// precheck file makes this test follow it; the two things worth locking in are
// that a single-stack cluster still passes every check, and that every
// disagreement between the node addresses and the cluster CIDRs is caught by
// exactly one check.
package core_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	kkprojectv1 "github.com/kubesphere/kubekey/api/project/v1"

	core "github.com/kubesphere/kubekey/v4/builtin/core"
	"github.com/kubesphere/kubekey/v4/pkg/converter/tmpl"
)

const precheckNetworkTasks = "roles/precheck/network/tasks/main.yaml"

// precheckCheck is one assertion of the precheck file: its name, the conditions
// that gate it, and the conditions it asserts.
type precheckCheck struct {
	name string
	when []string
	that []string
}

// loadPrecheckChecks flattens the task file into the assertions it performs.
// Blocks contribute their "when" to every task they contain; tasks without an
// "assert" are not assertions and are skipped.
func loadPrecheckChecks(t *testing.T) []precheckCheck {
	t.Helper()

	raw, err := core.BuiltinPlaybook.ReadFile(precheckNetworkTasks)
	require.NoError(t, err, "read %s", precheckNetworkTasks)

	var tasks []map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &tasks), "unmarshal %s", precheckNetworkTasks)

	var checks []precheckCheck
	for _, task := range tasks {
		if block, ok := task["block"].([]any); ok {
			outer := stringList(task["when"])
			for _, item := range block {
				sub, _ := item.(map[string]any)
				if sub == nil {
					continue
				}
				checks = append(checks, newPrecheckCheck(sub, outer)...)
			}
			continue
		}
		checks = append(checks, newPrecheckCheck(task, nil)...)
	}

	require.NotEmpty(t, checks, "%s contains no assertion", precheckNetworkTasks)
	return checks
}

func newPrecheckCheck(task map[string]any, extraWhen []string) []precheckCheck {
	assertArgs, _ := task["assert"].(map[string]any)
	if assertArgs == nil {
		return nil
	}
	name, _ := task["name"].(string)
	return []precheckCheck{{
		name: name,
		when: append(append([]string{}, extraWhen...), stringList(task["when"])...),
		that: stringList(assertArgs["that"]),
	}}
}

// stringList reads a value that the playbook schema allows as either a single
// string or a list of them.
func stringList(v any) []string {
	switch value := v.(type) {
	case string:
		return []string{value}
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// evalCondition mirrors pkg/modules/assert: the expression is wrapped in the
// template syntax first, then evaluated as a boolean.
func evalCondition(t *testing.T, cond string, vars map[string]any) bool {
	t.Helper()

	if !kkprojectv1.IsTmplSyntax(cond) {
		cond = kkprojectv1.ParseTmplSyntax(cond)
	}
	ok, err := tmpl.ParseBool(vars, cond)
	require.NoError(t, err, "evaluating %q", cond)

	return ok
}

// failingChecks returns the names of the checks whose conditions hold but whose
// assertion does not, in file order.
func failingChecks(t *testing.T, checks []precheckCheck, vars map[string]any) []string {
	t.Helper()

	var failed []string
	for _, check := range checks {
		skipped := false
		for _, when := range check.when {
			if !evalCondition(t, when, vars) {
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}
		for _, that := range check.that {
			if !evalCondition(t, that, vars) {
				failed = append(failed, check.name)
				break
			}
		}
	}

	return failed
}

func TestPrecheckNetworkAddressFamilies(t *testing.T) {
	checks := loadPrecheckChecks(t)

	cases := []struct {
		name         string
		podCIDR      string
		serviceCIDR  string
		internalIPv4 string
		internalIPv6 string
		// wantFailed are the checks expected to fail, in file order.
		wantFailed []string
	}{
		{
			// The out of the box configuration has to keep passing: one CIDR per
			// family group, one node address.
			name:         "IPv4-only, stock defaults",
			podCIDR:      "10.233.64.0/18",
			serviceCIDR:  "10.233.0.0/18",
			internalIPv4: "172.16.0.3",
		},
		{
			name:         "IPv4-only, custom range",
			podCIDR:      "172.16.0.0/16",
			serviceCIDR:  "172.26.0.0/16",
			internalIPv4: "192.168.0.3",
		},
		{
			name:         "IPv6-only",
			podCIDR:      "fd00::/64",
			serviceCIDR:  "fd00:10:96::/112",
			internalIPv6: "fd85::3",
		},
		{
			name:         "dual-stack",
			podCIDR:      "10.233.64.0/18,fd00::/64",
			serviceCIDR:  "10.233.0.0/18,fd00:10:96::/112",
			internalIPv4: "172.16.0.3",
			internalIPv6: "fd85::3",
		},
		{
			name:         "dual-stack cluster, IPv4-only node",
			podCIDR:      "10.233.64.0/18,fd00::/64",
			serviceCIDR:  "10.233.0.0/18,fd00:10:96::/112",
			internalIPv4: "172.16.0.3",
			wantFailed:   []string{"Network | Require both IPv4 and IPv6 node addresses on a dual-stack cluster"},
		},
		{
			name:         "dual-stack cluster, IPv6-only node",
			podCIDR:      "10.233.64.0/18,fd00::/64",
			serviceCIDR:  "10.233.0.0/18,fd00:10:96::/112",
			internalIPv6: "fd85::3",
			wantFailed:   []string{"Network | Require both IPv4 and IPv6 node addresses on a dual-stack cluster"},
		},
		{
			// The address used to be carried but silently unused for cluster
			// networking; it is now rejected with both remedies named.
			name:         "IPv4-only cluster, node also has an IPv6 address",
			podCIDR:      "10.233.64.0/18",
			serviceCIDR:  "10.233.0.0/18",
			internalIPv4: "172.16.0.3",
			internalIPv6: "fd85::3",
			wantFailed:   []string{"Network | Require IPv4-only node addresses on an IPv4-only cluster"},
		},
		{
			name:         "IPv6-only cluster, node also has an IPv4 address",
			podCIDR:      "fd00::/64",
			serviceCIDR:  "fd00:10:96::/112",
			internalIPv4: "172.16.0.3",
			internalIPv6: "fd85::3",
			wantFailed:   []string{"Network | Require IPv6-only node addresses on an IPv6-only cluster"},
		},
		{
			name:         "dual-stack pod CIDR, IPv4-only service CIDR",
			podCIDR:      "10.233.64.0/18,fd00::/64",
			serviceCIDR:  "10.233.0.0/18",
			internalIPv4: "172.16.0.3",
			internalIPv6: "fd85::3",
			wantFailed:   []string{"Network | Ensure pod CIDR and service CIDR use the same address families"},
		},
		{
			name:        "no node address at all",
			podCIDR:     "10.233.64.0/18",
			serviceCIDR: "10.233.0.0/18",
			wantFailed: []string{
				"Network | Ensure either internal_ipv4 or internal_ipv6 is defined",
				"Network | Require IPv4-only node addresses on an IPv4-only cluster",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vars := map[string]any{
				"cni": map[string]any{
					"type":           "calico",
					"pod_cidr":       tc.podCIDR,
					"service_cidr":   tc.serviceCIDR,
					"ipv4_mask_size": 24,
					"ipv6_mask_size": 64,
				},
				"kubernetes": map[string]any{
					"kube_version": "v1.33.5",
					"kubelet":      map[string]any{"max_pods": 110},
				},
				"cluster_require": map[string]any{
					"require_network_plugin": []any{"calico", "cilium", "flannel", "kubeovn"},
				},
				"internal_ipv4":      tc.internalIPv4,
				"internal_ipv6":      tc.internalIPv6,
				"inventory_hostname": "node1",
				"groups":             map[string]any{"k8s_cluster": []any{"node1"}},
			}

			assert.Equal(t, tc.wantFailed, failingChecks(t, checks, vars), "failing checks")
		})
	}
}
