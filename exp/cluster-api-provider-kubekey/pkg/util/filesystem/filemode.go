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
	"strconv"
)

type FileMode struct {
	os.FileMode
}

func (m FileMode) PermNumberString() string {
	const str = "dalTLDpSugct?"
	var buf [32]byte // Mode is uint32.
	w := 0
	for i, c := range str {
		if m.FileMode.Perm()&(1<<uint(32-1-i)) != 0 {
			buf[w] = byte(c)
			w++
		}
	}
	if w == 0 {
		buf[w] = '-'
		w++
	}
	const rwx = "421421421"
	res := ""
	cur := 0
	for i, c := range rwx {
		if (9-1-i)%3 == 0 {
			res = res + strconv.Itoa(cur)
			cur = 0
		}

		if m.FileMode.Perm()&(1<<uint(9-1-i)) != 0 {
			cnum, _ := strconv.Atoi(string(c))
			cur += cnum
		}
	}
	return res
}
