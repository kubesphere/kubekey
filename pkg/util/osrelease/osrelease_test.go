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

package osrelease

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

var testResult = []struct {
	name string
	data *Data
}{
	{
		name: "ubuntu2204",
		data: &Data{
			ID:         "ubuntu",
			IDLike:     "debian",
			Name:       "Ubuntu",
			PrettyName: "Ubuntu 22.04 LTS",
			Version:    "22.04 (Jammy Jellyfish)",
			VersionID:  "22.04",
		},
	},
	{
		name: "ubuntu2004",
		data: &Data{
			ID:         "ubuntu",
			IDLike:     "debian",
			Name:       "Ubuntu",
			PrettyName: "Ubuntu 20.04.4 LTS",
			Version:    "20.04.4 LTS (Focal Fossa)",
			VersionID:  "20.04",
		},
	},
	{
		name: "ubuntu1804",
		data: &Data{
			ID:         "ubuntu",
			IDLike:     "debian",
			Name:       "Ubuntu",
			PrettyName: "Ubuntu 18.04.2 LTS",
			Version:    "18.04.2 LTS (Bionic Beaver)",
			VersionID:  "18.04",
		},
	},
	{
		name: "centos7",
		data: &Data{
			ID:         "centos",
			IDLike:     "rhel fedora",
			Name:       "CentOS Linux",
			PrettyName: "CentOS Linux 7 (Core)",
			Version:    "7 (Core)",
			VersionID:  "7",
		},
	},
}

func TestParse(t *testing.T) {
	for _, tt := range testResult {
		t.Run(tt.name, func(t *testing.T) {
			bs, err := os.ReadFile(filepath.Join("test", tt.name))
			if err != nil {
				t.Errorf("cannot read datafile: %v", err)
			}
			info := Parse(string(bs))
			if !reflect.DeepEqual(tt.data, info) {
				t.Errorf("Parse() got = %v, want %v", info, tt.data)
			}
		})
	}
}

func TestParseByCommand(t *testing.T) {
	for _, tt := range testResult {
		t.Run(tt.name, func(t *testing.T) {
			c := fmt.Sprintf("cat %s", filepath.Join("test", tt.name))
			bs, err := exec.Command("/bin/bash", "-c", c).Output() //nolint:gosec
			if err != nil {
				t.Errorf("cannot cat datafile: %v", err)
			}
			info := Parse(string(bs))
			if !reflect.DeepEqual(tt.data, info) {
				t.Errorf("Parse() got = %v, want %v", info, tt.data)
			}
		})
	}
}
