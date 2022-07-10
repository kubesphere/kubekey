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
	"testing"
)

func TestFileMode_PermNumberString(t *testing.T) {
	tests := []struct {
		FileMode os.FileMode
		want     string
	}{
		{
			os.FileMode(0000),
			"000",
		},
		{
			os.FileMode(0777),
			"777",
		},
		{
			os.FileMode(0660),
			"660",
		},
		{
			os.FileMode(0755),
			"755",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			m := FileMode{
				FileMode: tt.FileMode,
			}
			if got := m.PermNumberString(); got != tt.want {
				t.Errorf("PermNumberString() = %v, want %v", got, tt.want)
			}
		})
	}
}
