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

package util

import (
	"testing"
	"time"
)

func TestShortDur(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{
			d:    3 * time.Second,
			want: "3s",
		},
		{
			d:    60 * time.Second,
			want: "1m",
		},
		{
			d:    4 * time.Minute,
			want: "4m",
		},
		{
			d:    1*time.Hour + 3*time.Second,
			want: "1h0m3s",
		},
		{
			d:    1*time.Hour + 4*time.Minute,
			want: "1h4m",
		},
		{
			d:    1*time.Hour + 4*time.Minute + 3*time.Second,
			want: "1h4m3s",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := ShortDur(tt.d); got != tt.want {
				t.Errorf("ShortDur() = %v, want %v", got, tt.want)
			}
		})
	}
}
