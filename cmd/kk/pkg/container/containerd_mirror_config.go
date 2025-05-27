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

package container

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/container/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/logger"
	"github.com/pkg/errors"
)

// GenerateContainerdMirrorConfig defines the action to generate containerd mirror config.
type GenerateContainerdMirrorConfig struct {
	common.KubeAction
}

// Execute creates config files for containerd registry mirrors.
func (g *GenerateContainerdMirrorConfig) Execute(runtime connector.Runtime) error {
	registry := g.KubeConf.Cluster.Registry

	if len(registry.RemoteMirrors) == 0 {
		return nil
	}

	// Create main certs.d directory
	certsDir := "/etc/containerd/certs.d"
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", certsDir), false); err != nil {
		return errors.Wrap(err, "failed to create certs.d directory")
	}

	// Create host.toml file for each registry in RemoteMirrors
	for registryHost, mirrorCfg := range registry.RemoteMirrors {
		// Create registry directory
		registryDir := filepath.Join(certsDir, registryHost)
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", registryDir), false); err != nil {
			return errors.Wrapf(err, "failed to create registry directory for %s", registryHost)
		}

		// Create hosts.toml file
		hostsTomlPath := filepath.Join(registryDir, "hosts.toml")
		data := templates.NewRemoteMirrorConfig(registryHost, mirrorCfg.Endpoint)

		var buffer bytes.Buffer
		tpl, err := templates.RegistryMirrors.Clone()
		if err != nil {
			return errors.Wrapf(err, "failed to clone template for %s", registryHost)
		}

		if err := tpl.Execute(&buffer, data); err != nil {
			return errors.Wrapf(err, "failed to render hosts.toml template for %s", registryHost)
		}
		renderContent := buffer.String()

		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cat > %s << EOF\n%s\nEOF", hostsTomlPath, renderContent), false); err != nil {
			return errors.Wrapf(err, "failed to create hosts.toml for %s", registryHost)
		}

		// Set permissions
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod 644 %s", hostsTomlPath), false); err != nil {
			return errors.Wrap(err, "failed to set permissions for hosts.toml")
		}

		// Log the creation
		logger.Log.Infof("Created mirror config for %s: %s", registryHost, hostsTomlPath)
	}

	return nil
}
