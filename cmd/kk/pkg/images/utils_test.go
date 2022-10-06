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

import (
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"reflect"
	"testing"
)

func TestParseArchVariant(t *testing.T) {
	type args struct {
		platform string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "test1",
			args: args{
				platform: "amd64",
			},
			want:  "amd64",
			want1: "",
		},
		{
			name: "test2",
			args: args{
				platform: "arm/v8",
			},
			want:  "arm",
			want1: "v8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseArchVariant(tt.args.platform)
			if got != tt.want {
				t.Errorf("ParseArchVariant() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseArchVariant() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseRepositoryTag(t *testing.T) {
	type args struct {
		repos string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
	}{
		{
			name: "test1",
			args: args{
				repos: "k8s.gcr.io/kube-apiserver-amd64:v1.16.3",
			},
			want:  "k8s.gcr.io/kube-apiserver-amd64",
			want1: "v1.16.3",
		},
		{
			name: "test2",
			args: args{
				repos: "docker.io/kubesphere/kube-apiserver:v1.16.3",
			},
			want:  "docker.io/kubesphere/kube-apiserver",
			want1: "v1.16.3",
		},
		{
			name: "test3",
			args: args{
				repos: "kube-apiserver:v1.16.3",
			},
			want:  "kube-apiserver",
			want1: "v1.16.3",
		},
		{
			name: "test4",
			args: args{
				repos: "calico:cni:v3.20.0-armv7",
			},
			want:  "calico:cni",
			want1: "v3.20.0-armv7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseImageTag(tt.args.repos)
			if got != tt.want {
				t.Errorf("ParseImageTag() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseImageTag() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestParseImageWithArchTag(t *testing.T) {
	tests := []struct {
		name  string
		ref   string
		want1 string
		want2 v1.Platform
	}{
		{
			name:  "t1",
			ref:   "kube-apiserver:v1.21.5-amd64",
			want1: "kube-apiserver:v1.21.5",
			want2: v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
				Variant:      "",
			},
		},
		{
			name:  "t1",
			ref:   "kube-apiserver:v1.21.5-amd64",
			want1: "kube-apiserver:v1.21.5",
			want2: v1.Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
		},
		{
			name:  "t2",
			ref:   "kube-apiserver:v1.21.5-arm-v7",
			want1: "kube-apiserver:v1.21.5",
			want2: v1.Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ParseImageWithArchTag(tt.ref)

			if got != tt.want1 {
				t.Errorf("ParseImageTag() got = %v, want %v", got, tt.want1)
			}

			if !reflect.DeepEqual(got1, tt.want2) {
				t.Errorf("ParseImageWithArchTag() = %v, want %v", got, tt.want2)
			}
		})
	}
}
