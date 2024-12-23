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

package builtin

import (
	"fmt"
	"os"

	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubesphere/kubekey/v4/builtin/core"
)

func completeInventory(inventoryFile string, inventory *kkcorev1.Inventory) error {
	if inventoryFile != "" {
		data, err := os.ReadFile(inventoryFile)
		if err != nil {
			return fmt.Errorf("cannot get inventory for inventoryFile: %s. error is %w.", inventoryFile, err)
		}
		return yaml.Unmarshal(data, inventory)
	}
	data, err := core.Defaults.ReadFile("defaults/inventory/localhost.yaml")
	if err != nil {
		return fmt.Errorf("cannot get local inventory. error is %w. Please set it by \"--inventory\"", err)
	}

	return yaml.Unmarshal(data, inventory)
}

func completeConfig(kubeVersion string, configFile string, config *kkcorev1.Config) error {
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("cannot get config for configFile: %s. error is %w.", kubeVersion, err)
		}
		return yaml.Unmarshal(data, config)
	}
	data, err := core.Defaults.ReadFile(fmt.Sprintf("defaults/config/%s.yaml", kubeVersion))
	if err != nil {
		return fmt.Errorf("cannot get config for kube_version: %s. error is %w. Please set it by \"--config\"", kubeVersion, err)
	}

	return yaml.Unmarshal(data, config)
}
