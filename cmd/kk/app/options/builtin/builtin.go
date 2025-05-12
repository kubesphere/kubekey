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

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubesphere/kubekey/v4/builtin/core"
)

const (
	defaultGroupControlPlane = "kube_control_plane"
	defaultGroupWorker       = "kube_worker"
)

func completeInventory(inventoryFile string, inventory *kkcorev1.Inventory) error {
	if inventoryFile != "" {
		data, err := os.ReadFile(inventoryFile)
		if err != nil {
			return errors.Wrapf(err, "failed to get inventory for inventoryFile: %q", inventoryFile)
		}
		return errors.Wrapf(yaml.Unmarshal(data, inventory), "failed to unmarshal inventoryFile %s", inventoryFile)
	}
	data, err := core.Defaults.ReadFile("defaults/inventory/localhost.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to get local inventory. Please set it by \"--inventory\"")
	}

	return errors.Wrapf(yaml.Unmarshal(data, inventory), "failed to unmarshal local inventoryFile %q", inventoryFile)
}

func completeConfig(kubeVersion string, configFile string, config *kkcorev1.Config) error {
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return errors.Wrapf(err, "failed to get configFile %q", configFile)
		}

		return errors.Wrapf(yaml.Unmarshal(data, config), "failed to unmarshal configFile %q", configFile)
	}
	data, err := core.Defaults.ReadFile(fmt.Sprintf("defaults/config/%s.yaml", kubeVersion))
	if err != nil {
		return errors.Wrapf(err, "failed to get local configFile for kube_version: %q. Please set it by \"--config\"", kubeVersion)
	}

	return errors.Wrapf(yaml.Unmarshal(data, config), "failed to unmarshal local configFile for kube_version: %q.", kubeVersion)
}
