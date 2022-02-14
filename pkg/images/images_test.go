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

package images

import "testing"

func Test_parseImageName(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		want  string
		want1 string
		want2 string
		want3 string
	}{
		{
			name:  "docker.io#calico#cni:v3.20.0.tar",
			file:  "docker.io#calico#cni:v3.20.0.tar",
			want:  "docker.io",
			want1: "calico",
			want2: "cni",
			want3: "v3.20.0",
		},
		{
			name:  "docker.io#kubesphere#kube-apiserver:v1.21.5.tar",
			file:  "docker.io#kubesphere#kube-apiserver:v1.21.5.tar",
			want:  "docker.io",
			want1: "kubesphere",
			want2: "kube-apiserver",
			want3: "v1.21.5",
		},
		{
			name:  "registry.cn-beijing.aliyuncs.com#kubesphere#kube-apiserver:v1.21.5.tar",
			file:  "registry.cn-beijing.aliyuncs.com#kubesphere#kube-apiserver:v1.21.5.tar",
			want:  "registry.cn-beijing.aliyuncs.com",
			want1: "kubesphere",
			want2: "kube-apiserver",
			want3: "v1.21.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3 := parseImageName(tt.file)
			if got != tt.want {
				t.Errorf("parseImageName() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseImageName() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("parseImageName() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.want3 {
				t.Errorf("parseImageName() got3 = %v, want %v", got3, tt.want3)
			}
		})
	}
}
