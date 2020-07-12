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

type Kubernetes struct {
	Version                string   `yaml:"version" json:"version,omitempty"`
	ImageRepo              string   `yaml:"imageRepo" json:"imageRepo,omitempty"`
	ClusterName            string   `yaml:"clusterName" json:"clusterName,omitempty"`
	MasqueradeAll          bool     `yaml:"masqueradeAll" json:"masqueradeAll,omitempty"`
	MaxPods                string   `yaml:"maxPods" json:"maxPods,omitempty"`
	NodeCidrMaskSize       string   `yaml:"nodeCidrMaskSize" json:"nodeCidrMaskSize,omitempty"`
	ApiserverCertExtraSans []string `yaml:"apiserverCertExtraSans" json:"apiserverCertExtraSans,omitempty"`
}
