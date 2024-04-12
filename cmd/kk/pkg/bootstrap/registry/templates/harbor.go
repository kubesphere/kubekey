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

package templates

import (
	"strings"
	"text/template"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/registry"

	"github.com/lithammer/dedent"
)

var (
	// HarborServiceTempl defines the template of registry's configuration file.
	HarborServiceTempl = template.Must(template.New("harborSerivce").Parse(
		dedent.Dedent(`[Unit]
Description=Harbor
After=docker.service systemd-networkd.service systemd-resolved.service
Requires=docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/docker-compose -f {{ .Harbor_install_path }}/docker-compose.yml up
ExecStop=/usr/local/bin/docker-compose -f {{ .Harbor_install_path }}/docker-compose.yml down
Restart=on-failure
[Install]
WantedBy=multi-user.target
    `)))
)

func Password(kubeConf *common.KubeConf, domain string) string {
	auths := registry.DockerRegistryAuthEntries(kubeConf.Cluster.Registry.Auths)
	for repo, entry := range auths {
		if strings.Contains(repo, domain) {
			return entry.Password
		}
	}

	return "Harbor12345"
}
