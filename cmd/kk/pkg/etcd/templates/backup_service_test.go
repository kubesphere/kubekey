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

package templates

import (
	"testing"
)

func TestBackupTimeOnCalendar(t *testing.T) {
	tests := []struct {
		period int
		want   string
	}{
		{
			30,
			"*-*-* *:00/30:00",
		},
		{
			60,
			"*-*-* 00/1:00:00",
		},
		{
			70,
			"*-*-* 00/1:10:00",
		},
		{
			1500,
			"*-*-* 00:00:00",
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := BackupTimeOnCalendar(tt.period); got != tt.want {
				t.Errorf("BackupTimeOnCalendar() = %v, want %v", got, tt.want)
			}
		})
	}
}
