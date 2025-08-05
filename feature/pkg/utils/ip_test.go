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

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIP(t *testing.T) {
	testcases := []struct {
		name     string
		ipRange  string
		excepted func() []string
	}{
		{
			name:    "parse cidr",
			ipRange: "192.168.1.0/30",
			excepted: func() []string {
				// 192.168.1.1 - 192.168.1.2
				return []string{"192.168.1.1", "192.168.1.2"}
			},
		},
		{
			name:    "parse single host cidr",
			ipRange: "10.0.0.1/32",
			excepted: func() []string {
				return []string{"10.0.0.1"}
			},
		},
		{
			name:    "parse range",
			ipRange: "192.168.1.1-192.168.1.3",
			excepted: func() []string {
				return []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
			},
		},
		{
			name:    "parse single ip",
			ipRange: "8.8.8.8",
			excepted: func() []string {
				return []string{"8.8.8.8"}
			},
		},
		{
			name:    "parse ipv6 cidr",
			ipRange: "2001:db8::/126",
			excepted: func() []string {
				return []string{"2001:db8::", "2001:db8::1", "2001:db8::2", "2001:db8::3"}
			},
		},
		{
			name:    "parse ipv6 range",
			ipRange: "2001:db8::1-2001:db8::3",
			excepted: func() []string {
				return []string{"2001:db8::1", "2001:db8::2", "2001:db8::3"}
			},
		},
		{
			name:    "parse ipv6 single",
			ipRange: "2001:db8::1",
			excepted: func() []string {
				return []string{"2001:db8::1"}
			},
		},
		{
			name:    "parse ip with mask",
			ipRange: "192.168.1.0/255.255.255.252",
			excepted: func() []string {
				return []string{"192.168.1.1", "192.168.1.2"}
			},
		},
		{
			name:    "invalid input",
			ipRange: "invalid",
			excepted: func() []string {
				return []string{"invalid"}
			},
		},
		{
			name:    "parse large cidr (truncated check)",
			ipRange: "192.168.0.0/18",
			excepted: func() []string {
				// 192.168.0.1 - 192.168.63.254
				var ips []string
				for i := range 64 {
					for j := range 256 {
						ips = append(ips, fmt.Sprintf("192.168.%d.%d", i, j))
					}
				}
				return ips[1 : len(ips)-1]
			},
		},
		{
			name:    "parse large range (truncated check)",
			ipRange: "192.168.0.1-192.168.63.254",
			excepted: func() []string {
				// 192.168.0.1 - 192.168.63.254
				var ips []string
				for i := range 64 {
					for j := range 256 {
						ips = append(ips, fmt.Sprintf("192.168.%d.%d", i, j))
					}
				}
				return ips[1 : len(ips)-1]
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted(), ParseIP(tc.ipRange))
		})
	}
}
