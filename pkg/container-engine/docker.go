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

package container_engine

import (
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	DockerServiceTempl = template.Must(template.New("DockerService").Parse(
		dedent.Dedent(`[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
# After=network-online.target firewalld.service containerd.service
# Wants=network-online.target
# Requires=docker.socket containerd.service

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always

# Note that StartLimit* options were moved from "Service" to "Unit" in systemd 229.
# Both the old, and new location are accepted by systemd 229 and up, so using the old location
# to make them work for either version of systemd.
StartLimitBurst=3

# Note that StartLimitInterval was renamed to StartLimitIntervalSec in systemd 230.
# Both the old, and new name are accepted by systemd 230 and up, so using the old name to make
# this option work for either version of systemd.
StartLimitInterval=60s

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Comment TasksMax if your systemd version does not support it.
# Only systemd 226 and above support this option.
TasksMax=infinity

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process
OOMScoreAdjust=-500

[Install]
WantedBy=multi-user.target

    `)))

	DockerConfigTempl = template.Must(template.New("DockerConfig").Parse(
		dedent.Dedent(`{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}],
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}],
  {{- end}}
  "exec-opts": ["native.cgroupdriver=systemd"]
}
    `)))
)

func generateDockerConfig(mgr *manager.Manager) (string, error) {
	var Mirrors, InsecureRegistries string
	if mgr.Cluster.Registry.RegistryMirrors != nil {
		mirrors := []string{}
		for _, mirror := range mgr.Cluster.Registry.RegistryMirrors {
			mirrors = append(mirrors, fmt.Sprintf("\"%s\"", mirror))
		}
		Mirrors = strings.Join(mirrors, ", ")
	}
	if mgr.Cluster.Registry.InsecureRegistries != nil {
		registries := []string{}
		for _, repostry := range mgr.Cluster.Registry.InsecureRegistries {
			registries = append(registries, fmt.Sprintf("\"%s\"", repostry))
		}
		InsecureRegistries = strings.Join(registries, ", ")
	}
	return util.Render(DockerConfigTempl, util.Data{
		"Mirrors":            Mirrors,
		"InsecureRegistries": InsecureRegistries,
	})
}

func generateDockerService() (string, error) {
	return util.Render(DockerServiceTempl, util.Data{})
}

func installDockerOnNode(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !manager.ExistNode(mgr, node) {
		output, _ := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"if [ -z $(which docker) ] || [ ! -e /var/run/docker.sock ]; then echo 'Container Runtime will be installed'; fi\"", 0, false)
		if strings.Contains(strings.TrimSpace(output), "Container Runtime will be installed") {
			err := syncDockerBinaries(mgr, node)
			if err != nil {
				return err
			}
			err = setContainerd(mgr)
			if err != nil {
				return err
			}
			err = setDocker(mgr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func syncDockerBinaries(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	tmpDir := kubekeyapiv1alpha1.DefaultTmpDir
	_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -d %s ]; then rm -rf %s ;fi\" && mkdir -p %s", tmpDir, tmpDir, tmpDir), 1, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create tmp dir")
	}

	currentDir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
	if err1 != nil {
		return errors.Wrap(err1, "Failed to get current dir")
	}

	filesDir := fmt.Sprintf("%s/%s/%s/%s", currentDir, kubekeyapiv1alpha1.DefaultPreDir, mgr.Cluster.Kubernetes.Version, node.Arch)

	docker := fmt.Sprintf("docker-%s.tgz", kubekeyapiv1alpha1.DefaultDockerVersion)

	if err := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filesDir, docker), fmt.Sprintf("%s/%s", "/tmp/kubekey", docker)); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to sync binaries"))
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /usr/bin && tar -zxf %s/%s && mv docker/* /usr/bin && rm -rf docker\"", "/tmp/kubekey", docker), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to install container runtime binaries"))
	}

	return nil
}

func setDocker(mgr *manager.Manager) error {
	// Generate systemd service for Docker
	dockerService, err := generateDockerService()
	if err != nil {
		return err
	}
	dockerServiceBase64 := base64.StdEncoding.EncodeToString([]byte(dockerService))
	_, err = mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/docker.service\"", dockerServiceBase64), 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate docker's service"))
	}

	// Generate daemon.json for Docker
	dockerConfig, err := generateDockerConfig(mgr)
	if err != nil {
		return err
	}
	dockerConfigBase64 := base64.StdEncoding.EncodeToString([]byte(dockerConfig))
	_, err = mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/docker && echo %s | base64 -d > /etc/docker/daemon.json\"", dockerConfigBase64), 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate docker's daemon.json"))
	}

	// Start Docker
	_, err = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl enable docker && systemctl start docker\"", 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to start docker"))
	}

	return nil
}
