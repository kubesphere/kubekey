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

package containermanager

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/v3/pkg/scope"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation/directory"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/v3/pkg/service/util"
)

// DockerService is a ContainerManager service implementation for docker.
type DockerService struct {
	sshClient     ssh.Interface
	scope         scope.KKInstanceScope
	instanceScope *scope.InstanceScope

	templateFactory   func(sshClient ssh.Interface, template *template.Template, data file.Data, dst string) (operation.Template, error)
	dockerFactory     func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	criDockerdFactory func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	crictlFactory     func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
}

// NewDockerService returns a new DockerService given the remote instance container manager client.
func NewDockerService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) *DockerService {
	return &DockerService{
		sshClient:     sshClient,
		scope:         scope,
		instanceScope: instanceScope,
	}
}

func (d *DockerService) getTemplateService(template *template.Template, data file.Data, dst string) (operation.Template, error) {
	if d.templateFactory != nil {
		return d.templateFactory(d.sshClient, template, data, dst)
	}
	return file.NewTemplate(d.sshClient, d.scope.RootFs(), template, data, dst)
}

func (d *DockerService) getDockerService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if d.dockerFactory != nil {
		return d.dockerFactory(sshClient, version, arch)
	}
	return file.NewDocker(sshClient, d.scope.RootFs(), version, arch)
}

func (d *DockerService) getCRIDockerdService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if d.criDockerdFactory != nil {
		return d.criDockerdFactory(sshClient, version, arch)
	}
	return file.NewCRIDockerd(sshClient, d.scope.RootFs(), version, arch)
}

func (d *DockerService) getCrictlService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if d.crictlFactory != nil {
		return d.crictlFactory(sshClient, version, arch)
	}
	return file.NewCrictl(sshClient, d.scope.RootFs(), version, arch)
}

// Type returns the type docker of the container manager.
func (d *DockerService) Type() string {
	return file.DockerID
}

// Version returns the version of the container manager.
func (d *DockerService) Version() string {
	return d.instanceScope.KKInstance.Spec.ContainerManager.Version
}

// CRIDockerdVersion returns the version of the cri-dockerd.
func (d *DockerService) CRIDockerdVersion() string {
	return d.instanceScope.KKInstance.Spec.ContainerManager.CRIDockerdVersion
}

// IsExist returns true if the container manager is installed.
func (d *DockerService) IsExist() bool {
	res, err := d.sshClient.SudoCmd(
		"if [ -z $(which docker) ] || [ ! -e /run/docker.sock ]; " +
			"then echo 'not exist'; " +
			"fi")
	if err != nil {
		return false
	}
	if strings.Contains(res, "not exist") {
		return false
	}
	return true
}

// Get gets the binary of docker and related components and copy them to the remote instance.
func (d *DockerService) Get(timeout time.Duration) error {
	docker, err := d.getDockerService(d.sshClient, d.Version(), d.instanceScope.Arch())
	if err != nil {
		return err
	}
	criDockerd, err := d.getCRIDockerdService(d.sshClient, d.CRIDockerdVersion(), d.instanceScope.Arch())
	if err != nil {
		return err
	}
	crictl, err := d.getCrictlService(d.sshClient, d.instanceScope.KKInstance.Spec.ContainerManager.CRICTLVersion, d.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		docker,
		criDockerd,
		crictl,
	}

	zone := d.scope.ComponentZone()
	host := d.scope.ComponentHost()
	overrideMap := make(map[string]infrav1.Override)
	for _, o := range d.scope.ComponentOverrides() {
		overrideMap[o.ID+o.Version+o.Arch] = o
	}

	for _, b := range binaries {
		override := overrideMap[b.ID()+b.Version()+b.Arch()]
		if err := util.DownloadAndCopy(d.instanceScope, b, zone, host, override.Path, override.URL, override.Checksum.Value, timeout); err != nil {
			return err
		}
	}

	// /usr/local
	dir := filepath.Dir(filepath.Dir(docker.RemotePath()))
	if _, err = d.sshClient.SudoCmdf("tar Cxzvf %s %s && mv %s/docker/* %s", dir, docker.RemotePath(), dir, directory.BinDir); err != nil {
		return err
	}
	dir = filepath.Dir(filepath.Dir(criDockerd.RemotePath()))
	if _, err = d.sshClient.SudoCmdf("tar Cxzvf %s %s && mv %s/cri-dockerd/* %s", dir, criDockerd.RemotePath(), dir, directory.BinDir); err != nil {
		return err
	}
	if _, err = d.sshClient.SudoCmdf("tar Cxzvf %s %s", filepath.Dir(crictl.RemotePath()), crictl.RemotePath()); err != nil {
		return err
	}
	return nil
}

// Install installs the container manager and related components.
func (d *DockerService) Install() error {
	if err := d.installDocker(); err != nil {
		return err
	}
	if err := d.installCRIDockerd(); err != nil {
		return err
	}
	if err := d.installCrictl(); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) installDocker() error {
	if err := d.generateDockerService(); err != nil {
		return err
	}
	if err := d.generateDockerConfig(); err != nil {
		return err
	}
	if _, err := d.sshClient.SudoCmd("systemctl daemon-reload && systemctl enable docker && systemctl start docker"); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) installCRIDockerd() error {
	if err := d.generateCRIDockerdService(); err != nil {
		return err
	}
	if err := d.generateCRIDockerdSocket(); err != nil {
		return err
	}
	if _, err := d.sshClient.SudoCmd("systemctl daemon-reload && systemctl enable cri-docker && systemctl enable --now cri-docker.socket && systemctl start cri-docker"); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) installCrictl() error {
	temp, err := template.ParseFS(f, "templates/crictl.yaml")
	if err != nil {
		return err
	}

	svc, err := d.getTemplateService(
		temp,
		file.Data{
			"Endpoint": d.instanceScope.ContainerManager().CRISocket,
		},
		filepath.Join("/etc/", temp.Name()))
	if err != nil {
		return err
	}
	if err := svc.RenderToLocal(); err != nil {
		return err
	}
	if err := svc.Copy(true); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) generateDockerService() error {
	temp, err := template.ParseFS(f, "templates/docker.service")
	if err != nil {
		return err
	}
	svc, err := d.getTemplateService(temp, nil, filepath.Join(file.SystemdDir, temp.Name()))
	if err != nil {
		return err
	}
	if err = svc.RenderToLocal(); err != nil {
		return err
	}
	if err = svc.Copy(true); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) generateDockerConfig() error {
	temp, err := template.ParseFS(f, "templates/daemon.json")
	if err != nil {
		return err
	}
	svc, err := d.getTemplateService(
		temp,
		file.Data{
			"Mirrors":            d.mirrors(),
			"InsecureRegistries": d.insecureRegistry(),
		},
		filepath.Join("/etc/docker", temp.Name()))
	if err != nil {
		return err
	}
	if err = svc.RenderToLocal(); err != nil {
		return err
	}
	if err = svc.Copy(true); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) generateCRIDockerdService() error {
	temp, err := template.ParseFS(f, "templates/cri-docker.service")
	if err != nil {
		return err
	}
	svc, err := d.getTemplateService(temp, nil, filepath.Join(file.SystemdDir, temp.Name()))
	if err != nil {
		return err
	}
	if err = svc.RenderToLocal(); err != nil {
		return err
	}
	if err = svc.Copy(true); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) generateCRIDockerdSocket() error {
	temp, err := template.ParseFS(f, "templates/cri-docker.socket")
	if err != nil {
		return err
	}
	svc, err := d.getTemplateService(temp, nil, filepath.Join(file.SystemdDir, temp.Name()))
	if err != nil {
		return err
	}
	if err = svc.RenderToLocal(); err != nil {
		return err
	}
	if err = svc.Copy(true); err != nil {
		return err
	}
	return nil
}

func (d *DockerService) mirrors() string {
	var m string
	if d.scope.GlobalRegistry() != nil {
		var mirrorsArr []string
		for _, mirror := range d.scope.GlobalRegistry().RegistryMirrors {
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("%q", mirror))
		}
		m = strings.Join(mirrorsArr, ", ")
	}
	return m
}

func (d *DockerService) insecureRegistry() string {
	var insecureRegistries string
	if d.scope.GlobalRegistry() != nil {
		var registriesArr []string
		for _, repo := range d.scope.GlobalRegistry().InsecureRegistries {
			registriesArr = append(registriesArr, fmt.Sprintf("%q", repo))
		}
		insecureRegistries = strings.Join(registriesArr, ", ")
	}
	return insecureRegistries
}
