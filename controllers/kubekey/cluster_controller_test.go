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

package kubekey

import "testing"

func Test_findIpAddress(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
		wantErr  bool
	}{
		{
			name:     "test_https",
			endpoint: "https://192.168.100.5:2379",
			want:     "192.168.100.5",
			wantErr:  false,
		},
		{
			name:     "test_http",
			endpoint: "http://192.168.100.5:2379",
			want:     "192.168.100.5",
			wantErr:  false,
		},
		{
			name:     "test_noport",
			endpoint: "https://192.168.100.5",
			want:     "192.168.100.5",
			wantErr:  false,
		},
		{
			name:     "test_domain",
			endpoint: "https://www.domain.com",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "test_errIp",
			endpoint: "https://192.168.100",
			want:     "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findIpAddress(tt.endpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("findIpAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findIpAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}
