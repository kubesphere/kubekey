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

package addons

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
)

type Install struct {
	common.KubeAction
}

func (i *Install) Execute(runtime connector.Runtime) error {
	nums := len(i.KubeConf.Cluster.Addons)

	enabledAddons, err := i.enabledAddons()
	if err != nil {
		return err
	}

	logger.Log.Messagef(runtime.RemoteHost().GetName(), "[%v/%v] enabled addons", len(enabledAddons), nums)

	for index, addon := range i.KubeConf.Cluster.Addons {
		if _, ok := enabledAddons[addon.Name]; !ok {
			continue
		}
		logger.Log.Messagef(runtime.RemoteHost().GetName(), "Install addon [%v-%v]: %s", nums, index, addon.Name)
		if err := InstallAddons(i.KubeConf, &addon, filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.GetObjName()))); err != nil {
			return err
		}
	}
	return nil
}

func (i *Install) enabledAddons() (map[string]struct{}, error) {
	enabledAddons := make(map[string]struct{}, len(i.KubeConf.Cluster.Addons))
	for _, addon := range i.KubeConf.Cluster.Addons {
		enabledAddons[addon.Name] = struct{}{}
	}

	if len(i.KubeConf.Arg.EnabledAddons) == 0 {
		return enabledAddons, nil
	}

	enabledAddonsConfig := make(map[string]struct{}, len(i.KubeConf.Arg.EnabledAddons))
	for _, config := range i.KubeConf.Arg.EnabledAddons {
		enabledAddonsConfig[config] = struct{}{}
	}

	for addon := range enabledAddons {
		if _, ok := enabledAddonsConfig[addon]; !ok {
			// drop addons not in input args
			delete(enabledAddons, addon)
		} else {
			// drop input args validated
			delete(enabledAddonsConfig, addon)
		}
	}
	if len(enabledAddonsConfig) > 0 {
		// exists invalid input args
		keys := make([]string, 0, len(enabledAddonsConfig))
		for key := range enabledAddonsConfig {
			keys = append(keys, key)
		}
		return nil, errors.New(fmt.Sprintf("Addons not exists: %s", strings.Join(keys, ",")))
	}
	return enabledAddons, nil
}
