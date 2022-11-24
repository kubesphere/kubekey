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
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/service/util"
)

// ContainerdService is a ContainerManager service implementation for containerd.
type ContainerdService struct {
	sshClient ssh.Interface

	scope         scope.KKInstanceScope
	instanceScope *scope.InstanceScope

	templateFactory   func(sshClient ssh.Interface, template *template.Template, data file.Data, dst string) (operation.Template, error)
	runcFactory       func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	containerdFactory func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	crictlFactory     func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
}

// NewContainerdService returns a new ContainerdService given the remote instance container manager client.
func NewContainerdService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) *ContainerdService {
	return &ContainerdService{
		sshClient:     sshClient,
		scope:         scope,
		instanceScope: instanceScope,
	}
}

func (s *ContainerdService) getRuncService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.runcFactory != nil {
		return s.runcFactory(sshClient, version, arch)
	}
	return file.NewRunc(sshClient, s.scope.RootFs(), version, arch)
}

func (s *ContainerdService) getContainerdService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.containerdFactory != nil {
		return s.containerdFactory(sshClient, version, arch)
	}
	return file.NewContainerd(sshClient, s.scope.RootFs(), version, arch)
}

func (s *ContainerdService) getCrictlService(sshClient ssh.Interface, version, arch string) (operation.Binary, error) {
	if s.crictlFactory != nil {
		return s.crictlFactory(sshClient, version, arch)
	}
	return file.NewCrictl(sshClient, s.scope.RootFs(), version, arch)
}

func (s *ContainerdService) getTemplateService(template *template.Template, data file.Data, dst string) (operation.Template, error) {
	if s.templateFactory != nil {
		return s.templateFactory(s.sshClient, template, data, dst)
	}
	return file.NewTemplate(s.sshClient, s.scope.RootFs(), template, data, dst)
}

// Type returns the type containerd of the container manager.
func (s *ContainerdService) Type() string {
	return file.ContainerdID
}

// Version returns the version of the container manager.
func (s *ContainerdService) Version() string {
	return s.instanceScope.KKInstance.Spec.ContainerManager.Version
}

// IsExist returns true if the container manager is installed.
func (s *ContainerdService) IsExist() bool {
	res, err := s.sshClient.SudoCmd(
		"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; " +
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

// Get gets the binary of containerd and related components and copy them to the remote instance.
func (s *ContainerdService) Get(timeout time.Duration) error {
	containerd, err := s.getContainerdService(s.sshClient, s.Version(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	runc, err := s.getRuncService(s.sshClient, file.RuncDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}
	crictl, err := s.getCrictlService(s.sshClient, s.instanceScope.KKInstance.Spec.ContainerManager.CRICTLVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		containerd,
		runc,
		crictl,
	}

	zone := s.scope.ComponentZone()
	host := s.scope.ComponentHost()
	overrideMap := make(map[string]infrav1.Override)
	for _, o := range s.scope.ComponentOverrides() {
		overrideMap[o.ID+o.Version+o.Arch] = o
	}

	for _, b := range binaries {
		override := overrideMap[b.ID()+b.Version()+b.Arch()]
		if err := util.DownloadAndCopy(s.instanceScope, b, zone, host, override.Path, override.URL, override.Checksum.Value, timeout); err != nil {
			return err
		}
	}

	// /usr/local
	dir := filepath.Dir(filepath.Dir(containerd.RemotePath()))
	if _, err := s.sshClient.SudoCmdf("tar Cxzvf %s %s", dir, containerd.RemotePath()); err != nil {
		return err
	}
	if _, err := s.sshClient.SudoCmdf("tar Cxzvf %s %s", filepath.Dir(crictl.RemotePath()), crictl.RemotePath()); err != nil {
		return err
	}
	return nil
}

// Install installs the container manager and related components.
func (s *ContainerdService) Install() error {
	if err := s.installContainerd(); err != nil {
		return err
	}
	if err := s.installRunc(); err != nil {
		return err
	}
	if err := s.installCrictl(); err != nil {
		return err
	}
	return nil
}

func (s *ContainerdService) installContainerd() error {
	if err := s.generateContainerdConfig(); err != nil {
		return err
	}
	if err := s.generateContainerdService(); err != nil {
		return err
	}
	if _, err := s.sshClient.SudoCmd("systemctl daemon-reload && systemctl enable containerd && systemctl start containerd"); err != nil {
		return err
	}
	return nil
}

func (s *ContainerdService) generateContainerdService() error {
	temp, err := template.ParseFS(f, "templates/containerd.service")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(temp, nil, filepath.Join(file.SystemdDir, temp.Name()))
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

func (s *ContainerdService) generateContainerdConfig() error {
	temp, err := template.ParseFS(f, "templates/config.toml")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(
		temp,
		file.Data{
			"Mirrors":            s.mirrors(),
			"InsecureRegistries": s.insecureRegistry(),
			// todo: handle sandbox image
			// "SandBoxImage":       images.GetImage(m.Runtime, m.KubeConf, "pause").ImageName(),
			"PrivateRegistry": s.privateRegistry(),
			"Auth":            s.auth(),
		},
		filepath.Join("/etc/containerd/", temp.Name()))
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

func (s *ContainerdService) mirrors() string {
	var m string
	if s.scope.GlobalRegistry() != nil {
		var mirrorsArr []string
		for _, mirror := range s.scope.GlobalRegistry().RegistryMirrors {
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("%q", mirror))
		}
		m = strings.Join(mirrorsArr, ", ")
	}
	return m
}

func (s *ContainerdService) insecureRegistry() []string {
	var insecureRegistries []string
	if s.scope.GlobalRegistry() != nil {
		insecureRegistries = s.scope.GlobalRegistry().InsecureRegistries
	}
	return insecureRegistries
}

func (s *ContainerdService) privateRegistry() string {
	if s.scope.GlobalRegistry() != nil {
		return s.scope.GlobalRegistry().PrivateRegistry
	}
	return ""
}

func (s *ContainerdService) auth() infrav1.RegistryAuth {
	if s.scope.GlobalRegistry() != nil {
		auth := s.scope.GlobalRegistry().Auth.DeepCopy()
		if auth.CertsPath != "" {
			ca, cert, key, err := s.lookupCertsFile(auth.CertsPath)
			if err != nil {
				s.scope.Info(fmt.Sprintf("Failed to lookup certs file from the specific cert path %s: %s", auth.CertsPath, err.Error()))
				return *auth
			}
			auth.CAFile = ca
			auth.CertsPath = cert
			auth.KeyFile = key
		}
		if auth.PlainHTTP {
			auth.InsecureSkipVerify = true
		}
		return *auth
	}
	return infrav1.RegistryAuth{}
}

func (s *ContainerdService) lookupCertsFile(path string) (ca string, cert string, key string, err error) {
	s.instanceScope.V(2).Info(fmt.Sprintf("Looking for TLS certificates and private keys in %s", path))
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	s.instanceScope.V(2).Info(fmt.Sprintf("Looking for TLS certificates and private keys in abs path %s", absPath))
	fs, err := os.ReadDir(absPath)
	if err != nil {
		return ca, cert, key, err
	}

	for _, f := range fs {
		fullPath := filepath.Join(path, f.Name())
		if strings.HasSuffix(f.Name(), ".crt") {
			s.instanceScope.V(2).Info(fmt.Sprintf(" crt: %s", fullPath))
			ca = fullPath
		}
		if strings.HasSuffix(f.Name(), ".cert") {
			certName := f.Name()
			keyName := certName[:len(certName)-5] + ".key"
			s.instanceScope.V(2).Info(fmt.Sprintf(" cert: %s", fullPath))
			if !hasFile(fs, keyName) {
				return ca, cert, key, errors.Errorf("missing key %s for client certificate %s. Note that CA certificates should use the extension .crt", keyName, certName)
			}
			cert = fullPath
		}
		if strings.HasSuffix(f.Name(), ".key") {
			keyName := f.Name()
			certName := keyName[:len(keyName)-4] + ".cert"
			s.instanceScope.V(2).Info(fmt.Sprintf(" key: %s", fullPath))
			if !hasFile(fs, certName) {
				return ca, cert, key, errors.Errorf("missing client certificate %s for key %s", certName, keyName)
			}
			key = fullPath
		}
	}
	return ca, cert, key, nil
}

func (s *ContainerdService) installRunc() error {
	runc, err := s.getRuncService(s.sshClient, file.RuncDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}

	if _, err := s.sshClient.SudoCmdf("install -m 755 %s /usr/local/sbin/runc", runc.RemotePath()); err != nil {
		return err
	}

	_, _ = s.sshClient.SudoCmdf("rm -rf %s", runc.RemotePath())
	return nil
}

func (s *ContainerdService) installCrictl() error {
	temp, err := template.ParseFS(f, "templates/crictl.yaml")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(
		temp,
		file.Data{
			"Endpoint": s.instanceScope.ContainerManager().CRISocket,
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

func hasFile(files []os.DirEntry, name string) bool {
	for _, f := range files {
		if f.Name() == name {
			return true
		}
	}
	return false
}
