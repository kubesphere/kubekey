/*
 Copyright 2021 The KubeSphere Authors.

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

package kubesphere

import (
	"github.com/kubesphere/kubekey/cmd/kk/pkg/version/kubesphere"
	"reflect"
	"testing"
)

func Test_mirrorRepo(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "test_latest",
			version: "latest",
			want:    "kubespheredev",
		},
		{
			name:    "test_master",
			version: "master",
			want:    "kubespheredev",
		},
		{
			name:    "test_v3.2.1-rc.1",
			version: "v3.2.1-rc.1",
			want:    "kubespheredev",
		},
		{
			name:    "test_v3.2.1",
			version: "v3.2.1",
			want:    "kubesphere",
		},
		{
			name:    "test_v3.2.0",
			version: "v3.2.0",
			want:    "kubesphere",
		},
		{
			name:    "test_3.2.0",
			version: "3.2.0",
			want:    "kubespheredev",
		},
		{
			name:    "test_v3.2.0-alpha.1",
			version: "v3.2.0-alpha.1",
			want:    "kubespheredev",
		},
		{
			name:    "test_v1.2.0",
			version: "v1.2.0",
			want:    "kubesphere",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo(tt.version)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StabledVersionSupport() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func repo(version string) string {
	var r string
	_, latest := kubesphere.LatestRelease(version)
	_, dev := kubesphere.DevRelease(version)
	_, stable := kubesphere.StabledVersionSupport(version)
	switch {
	case stable:
		r = "kubesphere"
	case dev:
		r = "kubespheredev"
	case latest:
		r = "kubespheredev"
	default:
		r = "kubesphere"
	}
	return r
}
