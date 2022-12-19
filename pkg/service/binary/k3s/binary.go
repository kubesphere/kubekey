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

package k3s

import (
	"embed"
	"path/filepath"
	"text/template"
	"time"

	infrav1 "github.com/kubesphere/kubekey/v3/api/v1beta1"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation"
	"github.com/kubesphere/kubekey/v3/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/v3/pkg/service/util"
)

//go:embed templates
var f embed.FS

// Download downloads binaries.
func (s *Service) Download(timeout time.Duration) error {
	if err := s.DownloadAll(timeout); err != nil {
		return err
	}
	if err := s.GenerateK3sInstallScript(); err != nil {
		return err
	}
	return nil
}

// DownloadAll downloads all binaries.
func (s *Service) DownloadAll(timeout time.Duration) error {
	k3s, err := s.getK3sService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubecni, err := s.getKubecniService(file.KubecniDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		k3s,
		kubecni,
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
		if err := b.Chmod("+x"); err != nil {
			return err
		}
	}

	if _, err := s.sshClient.SudoCmdf("tar Cxzvf %s %s", filepath.Dir(kubecni.RemotePath()), kubecni.RemotePath()); err != nil {
		return err
	}

	return nil
}

// GenerateK3sInstallScript generates k3s install script.
func (s *Service) GenerateK3sInstallScript() error {
	temp, err := template.ParseFS(f, "templates/k3s-install.sh")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(
		temp,
		nil,
		filepath.Join(file.BinDir, temp.Name()))
	if err != nil {
		return err
	}
	if err := svc.RenderToLocal(); err != nil {
		return err
	}
	if err := svc.Copy(true); err != nil {
		return err
	}
	if err := svc.Chmod("+x"); err != nil {
		return err
	}
	return nil
}

// UpgradeDownload downloads upgrade binaries.
func (s *Service) UpgradeDownload(timeout time.Duration) error {
	return s.downloadUpgradeBinaries(timeout)
}

// downloadUpgradeBinaries downloads k3s binary.
func (s *Service) downloadUpgradeBinaries(timeout time.Duration) error {
	k3s, err := s.getK3sService(s.instanceScope.InPlaceUpgradeVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		k3s,
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
		if err := b.Chmod("+x"); err != nil {
			return err
		}
	}

	return nil
}
