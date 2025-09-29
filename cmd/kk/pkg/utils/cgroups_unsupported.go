//go:build !linux
// +build !linux

package utils

func CgroupsV2() bool {
	return false
}
