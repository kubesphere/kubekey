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

package directory

import (
	"os"
	"testing"

	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/util/filesystem"
)

func Test_checkFileMode(t *testing.T) {
	tests := []struct {
		mode os.FileMode
		want filesystem.FileMode
	}{
		{
			0,
			filesystem.FileMode{FileMode: os.ModeDir | os.FileMode(0644)},
		},
		{
			os.FileMode(0664),
			filesystem.FileMode{FileMode: os.ModeDir | os.FileMode(0664)},
		},
		{
			os.FileMode(0777),
			filesystem.FileMode{FileMode: os.ModeDir | os.FileMode(0777)},
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := checkFileMode(tt.mode); got != tt.want {
				t.Errorf("checkFileMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
