//go:build !linux

package os

// shouldUnmountCalicoGroup checks if we should unmount calico cgroup
// On non-Linux systems, this is never needed
func shouldUnmountCalicoGroup() bool {
	return false
}
