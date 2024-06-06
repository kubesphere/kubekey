//go:build builtin
// +build builtin

/*
Copyright 2024 The KubeSphere Authors.

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

package options

import (
	"gopkg.in/yaml.v3"

	"github.com/kubesphere/kubekey/v4/builtin"
)

func init() {
	if err := yaml.Unmarshal(builtin.DefaultConfig, defaultConfig); err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(builtin.DefaultInventory, defaultInventory); err != nil {
		panic(err)
	}
}
