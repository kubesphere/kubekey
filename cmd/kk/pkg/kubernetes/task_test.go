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

package kubernetes

import (
	"testing"
)

func Test_calculateNextStr(t *testing.T) {
	tests := []struct {
		currentVersion string
		desiredVersion string
		want           string
		wantErr        bool
		errMsg         string
	}{
		{
			currentVersion: "v1.21.5",
			desiredVersion: "v1.22.5",
			want:           "v1.22.5",
			wantErr:        false,
		},
		{
			currentVersion: "v1.21.5",
			desiredVersion: "v1.23.5",
			want:           "v1.22.12",
			wantErr:        false,
		},
		{
			currentVersion: "v1.17.5",
			desiredVersion: "v1.18.5",
			want:           "",
			wantErr:        true,
			errMsg:         "the target version v1.18.5 is not supported",
		},
		{
			currentVersion: "v1.17.5",
			desiredVersion: "v1.21.5",
			want:           "",
			wantErr:        true,
			errMsg:         "Kubernetes minor version v1.18.x is not supported",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got, err := calculateNextStr(tt.currentVersion, tt.desiredVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateNextStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calculateNextStr() got = %v, want %v", got, tt.want)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("calculateNextStr() error = %v, want %v", err, tt.errMsg)
			}
		})
	}
}
