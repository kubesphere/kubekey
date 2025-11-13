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
	"bytes"
	"fmt"
	"text/template"

	"github.com/cockroachdb/errors"
	kkcorev1 "github.com/kubesphere/kubekey/api/core/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/kubesphere/kubekey/v4/builtin/core"
	"github.com/kubesphere/kubekey/v4/cmd/kk/app/options"
)

const (
	defaultKubeVersion = "v1.33.1"
)

const (
	defaultGroupControlPlane = "kube_control_plane"
	defaultGroupWorker       = "kube_worker"
)

var getInventory options.InventoryFunc = func() (*kkcorev1.Inventory, error) {
	data, err := core.Defaults.ReadFile("defaults/inventory/localhost.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get local inventory. Please set it by \"--inventory\"")
	}

	inventory := &kkcorev1.Inventory{}
	return inventory, errors.Wrap(yaml.Unmarshal(data, inventory), "failed to unmarshal local inventory file: %q.")
}

func getInventoryData() ([]byte, error) {
	return core.Defaults.ReadFile("defaults/inventory/localhost.yaml")
}

func getConfig(kubeVersion string) ([]byte, error) {
	t, err := template.ParseFS(core.Defaults, fmt.Sprintf("defaults/config/%s.yaml", kubeVersion[:5]))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get local configFile template for kube_version: %q. Please set it by \"--config\"", kubeVersion)
	}
	data := bytes.NewBuffer(nil)
	if err := t.Execute(data, map[string]string{"kube_version": kubeVersion}); err != nil {
		return nil, errors.Wrapf(err, "failed to parse local configFile template for kube_version: %q. Please set it by \"--config\"", kubeVersion)
	}

	return data.Bytes(), nil
}
