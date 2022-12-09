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

package kubernetes

import (
	"embed"
	"path/filepath"
	"text/template"
	"time"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/service/operation"
	"github.com/kubesphere/kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/service/util"
)

//go:embed templates
var f embed.FS

// Download downloads binaries.
func (s *Service) Download(timeout time.Duration) error {
	if err := s.DownloadAll(timeout); err != nil {
		return err
	}

	if err := s.ConfigureKubelet(); err != nil {
		return err
	}

	if err := s.ConfigureKubeadm(); err != nil {
		return err
	}

	return nil
}

// DownloadAll downloads all binaries.
func (s *Service) DownloadAll(timeout time.Duration) error {
	kubeadm, err := s.getKubeadmService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubelet, err := s.getKubeletService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubecni, err := s.getKubecniService(file.KubecniDefaultVersion, s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubectl, err := s.getKubectlService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		kubeadm,
		kubelet,
		kubecni,
		kubectl,
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

// ConfigureKubelet configures kubelet.
func (s *Service) ConfigureKubelet() error {
	kubelet, err := s.getKubeletService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	if _, err := s.sshClient.SudoCmdf("ln -snf %s /usr/bin/kubelet", kubelet.RemotePath()); err != nil {
		return err
	}

	temp, err := template.ParseFS(f, "templates/kubelet.service")
	if err != nil {
		return err
	}

	svc, err := s.getTemplateService(
		temp,
		nil,
		filepath.Join(file.SystemdDir, temp.Name()))
	if err != nil {
		return err
	}
	if err := svc.RenderToLocal(); err != nil {
		return err
	}
	if err := svc.Copy(true); err != nil {
		return err
	}

	env, err := template.ParseFS(f, "templates/kubelet.conf")
	if err != nil {
		return err
	}

	envSvc, err := s.getTemplateService(
		env,
		file.Data{
			"NodeIP":   s.instanceScope.InternalAddress(),
			"Hostname": s.instanceScope.HostName(),
		},
		filepath.Join(file.SystemdDir, "kubelet.service.d", env.Name()),
	)
	if err != nil {
		return err
	}
	if err := envSvc.RenderToLocal(); err != nil {
		return err
	}
	if err := envSvc.Copy(true); err != nil {
		return err
	}

	if _, err := s.sshClient.SudoCmdf("systemctl disable kubelet && systemctl enable kubelet"); err != nil {
		return err
	}
	return nil
}

// ConfigureKubeadm configures kubeadm.
func (s *Service) ConfigureKubeadm() error {
	kubeadm, err := s.getKubeadmService(s.instanceScope.KubernetesVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	if _, err := s.sshClient.SudoCmdf("ln -snf %s /usr/bin/kubeadm", kubeadm.RemotePath()); err != nil {
		return err
	}
	return nil
}

// UpgradeDownload downloads upgrade binaries.
func (s *Service) UpgradeDownload(timeout time.Duration) error {
	return s.downloadUpgradeBinaries(timeout)
}

// downloadUpgradeBinaries downloads kubeadm, kubelet and kubectl.
func (s *Service) downloadUpgradeBinaries(timeout time.Duration) error {
	kubeadm, err := s.getKubeadmService(s.instanceScope.InPlaceUpgradeVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubelet, err := s.getKubeletService(s.instanceScope.InPlaceUpgradeVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}
	kubectl, err := s.getKubectlService(s.instanceScope.InPlaceUpgradeVersion(), s.instanceScope.Arch())
	if err != nil {
		return err
	}

	binaries := []operation.Binary{
		kubeadm,
		kubelet,
		kubectl,
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
		if _, err := s.sshClient.SudoCmdf("ln -snf %s /usr/bin/%s", b.RemotePath(), b.Name()); err != nil {
			return err
		}
	}

	return nil
}
