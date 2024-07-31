package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIp(t *testing.T) {
	testcases := []struct {
		name     string
		ipRange  string
		excepted func() []string
	}{
		{
			name:    "parse cidr",
			ipRange: "192.168.0.0/18",
			excepted: func() []string {
				// 192.168.0.1 - 192.168.63.254
				var ips []string
				for i := 0; i <= 63; i++ {
					for j := 0; j <= 255; j++ {
						ips = append(ips, fmt.Sprintf("192.168.%d.%d", i, j))
					}
				}
				return ips[1 : len(ips)-1]
			},
		},
		{
			name:    "parse range",
			ipRange: "192.168.0.1-192.168.63.254",
			excepted: func() []string {
				// 192.168.0.1 - 192.168.63.254
				var ips []string
				for i := 0; i <= 63; i++ {
					for j := 0; j <= 255; j++ {
						ips = append(ips, fmt.Sprintf("192.168.%d.%d", i, j))
					}
				}
				return ips[1 : len(ips)-1]
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted(), parseIp(tc.ipRange))
		})
	}
}
