//go:build linux
// +build linux

package utils

import (
	"github.com/containerd/cgroups/v3"
)

func CgroupsV2() bool {
	if cgroups.Mode() == cgroups.Unified || cgroups.Mode() == cgroups.Hybrid {
		return true
	}
	return false
}
