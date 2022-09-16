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

package v1beta1

// ISO type
const (
	NONE = "none"
	AUTO = "auto"
)

// Repository defines the repository of the instance.
type Repository struct {
	// ISO specifies the ISO file name. There are 3 options:
	// "": empty string means will not install the packages.
	// "none": no ISO file will be used. And capkk will use the default repository to install the required packages.
	// "auto": capkk will detect the ISO file automatically. Only support Ubuntu/Debian/CentOS.
	// "xxx-20.04-debs-amd64.iso": use the specified name to get the ISO file name.
	// +optional
	ISO string `json:"iso,omitempty"`

	// Update will update the repository packages list and cache if it is true.
	// +optional
	Update bool `json:"update,omitempty"`

	// Packages is a list of packages to be installed.
	// +optional
	Packages []string `json:"packages,omitempty"`
}
