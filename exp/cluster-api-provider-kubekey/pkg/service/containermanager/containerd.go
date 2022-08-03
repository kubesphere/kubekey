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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"

	infrav1 "github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/clients/ssh"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/scope"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/exp/cluster-api-provider-kubekey/pkg/service/operation/file"
)

type ContainerdService struct {
	SSHClient ssh.Interface

	scope         scope.KKInstanceScope
	instanceScope *scope.InstanceScope

	templateFactory   func(sshClient ssh.Interface, template *template.Template, data file.Data, dst string) (operation.Template, error)
	runcFactory       func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	containerdFactory func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
	crictlFactory     func(sshClient ssh.Interface, version, arch string) (operation.Binary, error)
}

func NewContainerdService(sshClient ssh.Interface, scope scope.KKInstanceScope, instanceScope *scope.InstanceScope) *ContainerdService {
	return &ContainerdService{
		SSHClient:     sshClient,
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
	if s.containerdFactory != nil {
		return s.crictlFactory(sshClient, version, arch)
	}
	return file.NewCrictl(sshClient, s.scope.RootFs(), version, arch)
}

func (s *ContainerdService) getTemplateService(template *template.Template, data file.Data, dst string) (operation.Template, error) {
	if s.templateFactory != nil {
		return s.templateFactory(s.SSHClient, template, data, dst)
	}
	return file.NewTemplate(s.SSHClient, s.scope.RootFs(), template, data, dst)
}

func (s *ContainerdService) Type() string {
	return file.ContainerdID
}

func (s *ContainerdService) Version() string {
	return s.instanceScope.KKInstance.Spec.ContainerManager.Version
}

func (s *ContainerdService) IsExist() bool {
	res, err := s.SSHClient.SudoCmd(
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

func (s *ContainerdService) Get(timeout time.Duration) error {
	containerd, err := s.getContainerdService(s.SSHClient, s.Version(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	runc, err := s.getRuncService(s.SSHClient, file.RuncDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}
	crictl, err := s.getCrictlService(s.SSHClient, getFirstMajorVersion(s.instanceScope.KubernetesVersion()), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	binaries := []operation.Binary{
		containerd,
		runc,
		crictl,
	}

	for _, b := range binaries {
		needGet := true
		if b.LocalExist() && b.CompareChecksum() == nil {
			needGet = false
		}
		if needGet {
			if err := b.Get(timeout); err != nil {
				return err
			}
			if err := b.CompareChecksum(); err != nil {
				return err
			}
		}
		if err := b.Copy(true); err != nil {
			return err
		}
	}

	// /usr/local
	dir := filepath.Dir(filepath.Dir(containerd.RemotePath()))
	if _, err := s.SSHClient.SudoCmdf("tar Cxzvf %s %s", dir, containerd.RemotePath()); err != nil {
		return err
	}
	if _, err := s.SSHClient.SudoCmdf("tar Cxzvf %s %s", filepath.Dir(crictl.RemotePath()), crictl.RemotePath()); err != nil {
		return err
	}
	return nil
}

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
	if _, err := s.SSHClient.SudoCmd("systemctl daemon-reload && systemctl enable containerd && systemctl start containerd"); err != nil {
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
			//"SandBoxImage":       images.GetImage(m.Runtime, m.KubeConf, "pause").ImageName(),
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
			mirrorsArr = append(mirrorsArr, fmt.Sprintf("\"%s\"", mirror))
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
	fs, err := ioutil.ReadDir(absPath)
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
	runc, err := s.getRuncService(s.SSHClient, file.RuncDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}

	if _, err := s.SSHClient.SudoCmdf("install -m 755 %s /usr/local/sbin/runc", runc.RemotePath()); err != nil {
		return err
	}

	_, _ = s.SSHClient.SudoCmdf("rm -rf %s", runc.RemotePath())
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

func hasFile(files []os.FileInfo, name string) bool {
	for _, f := range files {
		if f.Name() == name {
			return true
		}
	}
	return false
}

func getFirstMajorVersion(version string) string {
	semantic := versionutil.MustParseSemantic(version)
	semantic = semantic.WithPatch(0)
	return fmt.Sprintf("v%s", semantic.String())
}
