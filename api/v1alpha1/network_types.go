/*
Copyright 2020 The KubeSphere Authors.

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

package v1alpha1

type NetworkConfig struct {
	Plugin          string    `yaml:"plugin" json:"plugin,omitempty"`
	KubePodsCIDR    string    `yaml:"kubePodsCIDR" json:"kubePodsCIDR,omitempty"`
	KubeServiceCIDR string    `yaml:"kubeServiceCIDR" json:"kubeServiceCIDR,omitempty"`
	Calico          CalicoCfg `yaml:"calico" json:"calico,omitempty"`
}

type CalicoCfg struct {
	IPIPMode  string `yaml:"ipipMode" json:"ipipMode,omitempty"`
	VXLANMode string `yaml:"vxlanMode" json:"vxlanMode,omitempty"`
	VethMTU   string `yaml:"vethMTU" json:"vethMTU,omitempty"`
}
