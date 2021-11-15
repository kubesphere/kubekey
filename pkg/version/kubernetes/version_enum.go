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

package kubernetes

import (
	"fmt"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type Version int

const (
	V115 Version = iota
	V116
	V117
	V118
	V119
	V120
	V121
	V122
)

var VersionList = []Version{
	V115,
	V116,
	V117,
	V118,
	V119,
	V120,
	V121,
	V122,
}

func (v Version) String() string {
	switch v {
	case V115:
		return "v1.15"
	case V116:
		return "v1.16"
	case V117:
		return "v1.17"
	case V118:
		return "v1.18"
	case V119:
		return "v1.19"
	case V120:
		return "v1.20"
	case V121:
		return "v1.21"
	case V122:
		return "v1.22"
	default:
		return "invalid option"
	}
}

func VersionSupport(version string) bool {
	K8sTargetVersion := versionutil.MustParseSemantic(version)
	for i := range VersionList {
		if VersionList[i].String() == fmt.Sprintf("v%v.%v", K8sTargetVersion.Major(), K8sTargetVersion.Minor()) {
			return true
		}
	}
	return false
}
