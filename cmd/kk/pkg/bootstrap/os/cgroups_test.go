package os

import (
	"runtime"
	"testing"
)

func TestShouldUnmountCalicoGroup(t *testing.T) {
	result := shouldUnmountCalicoGroup()
	
	// On Linux, the result depends on the actual cgroups mode
	// On non-Linux platforms, it should always return false
	if runtime.GOOS != "linux" {
		if result {
			t.Errorf("shouldUnmountCalicoGroup() should return false on non-Linux platforms, got %v", result)
		}
	}
	
	// The function should not panic regardless of platform
	t.Logf("shouldUnmountCalicoGroup() returned %v on %s", result, runtime.GOOS)
}
