/*
Copyright 2023 The KubeSphere Authors.

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

package connector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertBytesToMap(t *testing.T) {
	testcases := []struct {
		name     string
		data     []byte
		excepted map[string]string
	}{
		{
			name: "succeed",
			data: []byte(`PRETTY_NAME="Ubuntu 22.04.1 LTS"
NAME="Ubuntu"
VERSION_ID="22.04"
VERSION="22.04.1 LTS (Jammy Jellyfish)"
VERSION_CODENAME=jammy
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=jammy
`),
			excepted: map[string]string{
				"PRETTY_NAME":        "\"Ubuntu 22.04.1 LTS\"",
				"NAME":               "\"Ubuntu\"",
				"VERSION_ID":         "\"22.04\"",
				"VERSION":            "\"22.04.1 LTS (Jammy Jellyfish)\"",
				"VERSION_CODENAME":   "jammy",
				"ID":                 "ubuntu",
				"ID_LIKE":            "debian",
				"HOME_URL":           "\"https://www.ubuntu.com/\"",
				"SUPPORT_URL":        "\"https://help.ubuntu.com/\"",
				"BUG_REPORT_URL":     "\"https://bugs.launchpad.net/ubuntu/\"",
				"PRIVACY_POLICY_URL": "\"https://www.ubuntu.com/legal/terms-and-policies/privacy-policy\"",
				"UBUNTU_CODENAME":    "jammy",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted, convertBytesToMap(tc.data, "="))
		})
	}
}

func TestConvertBytesToSlice(t *testing.T) {
	testcases := []struct {
		name     string
		data     []byte
		excepted []map[string]string
	}{
		{
			name: "succeed",
			data: []byte(`processor	: 0
vendor_id	: GenuineIntel
cpu family	: 6
model		: 60
model name	: Intel Core Processor (Haswell, no TSX, IBRS)

processor	: 1
vendor_id	: GenuineIntel
cpu family	: 6
`),
			excepted: []map[string]string{
				{
					"processor":  "0",
					"vendor_id":  "GenuineIntel",
					"cpu family": "6",
					"model":      "60",
					"model name": "Intel Core Processor (Haswell, no TSX, IBRS)",
				},
				{
					"processor":  "1",
					"vendor_id":  "GenuineIntel",
					"cpu family": "6",
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.excepted, convertBytesToSlice(tc.data, ":"))
		})
	}
}
