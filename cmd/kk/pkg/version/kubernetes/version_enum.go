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
	"sort"

	versionutil "k8s.io/apimachinery/pkg/util/version"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
)

type Version int

const (
	V119 Version = iota
	V120
	V121
	V122
	V123
	V124
	V125
)

var VersionList = []Version{
	V119,
	V120,
	V121,
	V122,
	V123,
	V124,
	V125,
}

func (v Version) String() string {
	switch v {
	case V119:
		return "v1.19"
	case V120:
		return "v1.20"
	case V121:
		return "v1.21"
	case V122:
		return "v1.22"
	case V123:
		return "v1.23"
	case V124:
		return "v1.24"
	case V125:
		return "v1.25"
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

// SupportedK8sVersionList returns the supported list of Kubernetes
func SupportedK8sVersionList() []string {

	versionsList := make([]string, 0, len(files.FileSha256["kubeadm"]["amd64"]))
	for k := range files.FileSha256["kubeadm"]["amd64"] {
		versionsList = append(versionsList, k)
	}
	sort.Slice(versionsList, func(i, j int) bool {
		return versionutil.MustParseSemantic(versionsList[i]).LessThan(versionutil.MustParseSemantic(versionsList[j]))
	})

	return versionsList
}
