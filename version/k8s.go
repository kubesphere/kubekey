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

package version

// SupportedK8sVersionList returns the supported list of Kubernetes
func SupportedK8sVersionList() []string {
	return []string{
		"v1.15.12",
		"v1.16.8",
		"v1.16.10",
		"v1.16.12",
		"v1.16.13",
		"v1.17.0",
		"v1.17.4",
		"v1.17.5",
		"v1.17.6",
		"v1.17.7",
		"v1.17.8",
		"v1.17.9",
		"v1.18.3",
		"v1.18.5",
		"v1.18.6",
		"v1.18.8",
		"v1.19.0",
		"v1.19.8",
		"v1.19.9",
		"v1.20.4",
		"v1.20.6",
		"v1.20.10",
		"v1.21.4",
		"v1.21.5",
		"v1.22.1",
	}
}
