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

import "strings"

type Version int

const (
	V211 Version = iota
	V300
	V310
	V311
	V320
)

var VersionList = []Version{
	V211,
	V300,
	V310,
	V311,
	V320,
}

var VersionMap = map[string]*KsInstaller{
	V211.String(): KsV211,
	V300.String(): KsV300,
	V310.String(): KsV310,
	V311.String(): KsV311,
	V320.String(): KsV320,
}

var CNSource = map[string]bool{
	V310.String(): true,
	V311.String(): true,
	V320.String(): true,
}

func VersionSupport(version string) bool {
	if _, ok := VersionMap[version]; ok {
		return true
	}
	return false
}

func PreRelease(version string) bool {
	if strings.HasPrefix(version, "nightly-") ||
		version == "latest" ||
		strings.Contains(version, "alpha") ||
		strings.Contains(version, "release") {
		return true
	}
	return false
}

func Latest() *KsInstaller {
	return VersionMap[VersionList[len(VersionList)-1].String()]
}

func (v Version) String() string {
	switch v {
	case V211:
		return "v2.1.1"
	case V300:
		return "v3.0.0"
	case V310:
		return "v3.1.0"
	case V311:
		return "v3.1.1"
	case V320:
		return "v3.2.0"
	default:
		return "invalid option"
	}
}
