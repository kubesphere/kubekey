/*
 Copyright 2022 The KubeSphere Authors.

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

package filesystem

import (
	"os"
	"syscall"
)

const (
	sIsuid = syscall.S_ISUID
	sIsgid = syscall.S_ISGID
	sIsvtx = syscall.S_ISVTX
)

// ToChmodPerm converts Go permission bits to POSIX permission bits.
//
// This differs from fromFileMode in that we preserve the POSIX versions of
// setuid, setgid and sticky in m, because we've historically supported those
// bits, and we mask off any non-permission bits.
func ToChmodPerm(m os.FileMode) (perm uint32) {
	const mask = os.ModePerm | sIsuid | sIsgid | sIsvtx
	perm = uint32(m & mask)

	if m&os.ModeSetuid != 0 {
		perm |= sIsuid
	}
	if m&os.ModeSetgid != 0 {
		perm |= sIsgid
	}
	if m&os.ModeSticky != 0 {
		perm |= sIsvtx
	}

	return perm
}
