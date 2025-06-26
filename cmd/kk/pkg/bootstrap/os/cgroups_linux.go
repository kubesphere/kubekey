//go:build linux

package os

import (
	"github.com/containerd/cgroups/v3"
)

// shouldUnmountCalicoGroup checks if we should unmount calico cgroup
// This is only relevant on Linux systems with unified or hybrid cgroups
func shouldUnmountCalicoGroup() bool {
	return cgroups.Mode() == cgroups.Unified || cgroups.Mode() == cgroups.Hybrid
}
