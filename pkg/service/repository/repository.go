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

package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	infrav1 "github.com/kubesphere/kubekey/api/v1beta1"
	"github.com/kubesphere/kubekey/pkg/service/operation/file"
	"github.com/kubesphere/kubekey/pkg/service/operation/repository"
	"github.com/kubesphere/kubekey/pkg/service/util"
	"github.com/kubesphere/kubekey/pkg/util/filesystem"
	"github.com/kubesphere/kubekey/util/osrelease"
)

// Check checks the OS release info.
func (s *Service) Check() error {
	if !s.instanceScope.RepositoryEnabled() {
		return nil
	}

	output, err := s.sshClient.SudoCmd("cat /etc/os-release")
	if err != nil {
		return errors.Wrap(err, "failed to get os release")
	}

	s.scope.V(4).Info("Get os release", "output", output)
	osrData := osrelease.Parse(output)
	if osrData == nil || osrData.ID == "" || osrData.VersionID == "" {
		return errors.Errorf("failed to parse os release: %s", output)
	}
	s.os = osrData
	return nil
}

// Get gets the binary of ISO file and copy it to the remote instance.
func (s *Service) Get(timeout time.Duration) error {
	if !s.instanceScope.RepositoryUseISO() {
		return nil
	}

	s.scope.V(4).Info("os release", "os", s.os)
	iso, err := s.getISOService(s.os, s.instanceScope.Arch(), s.instanceScope.Repository().ISO)
	if err != nil {
		return err
	}

	zone := s.scope.ComponentZone()
	host := s.scope.ComponentHost()
	overrideMap := make(map[string]infrav1.Override)
	for _, o := range s.scope.ComponentOverrides() {
		overrideMap[o.ID+o.Version+o.Arch] = o
	}

	override := overrideMap[iso.ID()+iso.Version()+iso.Arch()]
	iso.HTTPChecksum.SetHost(host)
	iso.HTTPChecksum.SetPath(override.Checksum.Path)
	if err := util.DownloadAndCopy(s.instanceScope, iso, zone, host, override.Path, override.URL, override.Checksum.Value, timeout); err != nil {
		return err
	}
	return nil
}

// MountISO mounts the ISO file to the remote instance.
func (s *Service) MountISO() error {
	if !s.instanceScope.RepositoryUseISO() {
		return nil
	}

	mountPath := filepath.Join(file.MntDir, "repository")
	dirSvc := s.getDirectoryService(mountPath, os.FileMode(filesystem.FileMode0755))
	if err := dirSvc.Make(); err != nil {
		return errors.Wrapf(err, "failed to make directory %s", mountPath)
	}

	iso, err := s.getISOService(s.os, s.instanceScope.Arch(), s.instanceScope.Repository().ISO)
	if err != nil {
		return err
	}

	if _, err := s.sshClient.SudoCmd(fmt.Sprintf("sudo mount -t iso9660 -o loop %s %s", iso.RemotePath(), mountPath)); err != nil {
		return errors.Wrapf(err, "mount %s at %s failed", iso.RemotePath(), mountPath)
	}
	s.mountPath = mountPath
	return nil
}

// UmountISO unmounts the ISO file from the remote instance.
func (s *Service) UmountISO() error {
	if !s.instanceScope.RepositoryUseISO() {
		return nil
	}

	if _, err := s.sshClient.SudoCmd(fmt.Sprintf("sudo umount %s", s.mountPath)); err != nil {
		return errors.Wrapf(err, "umount %s failed", s.mountPath)
	}
	return nil
}

// UpdateAndInstall updates the linux package manager and installs some tools.
// Ex:
// apt-get update && apt-get install -y socat conntrack ipset ebtables chrony ipvsadm
// yum clean all && yum makecache && yum install -y openssl socat conntrack ipset ebtables chrony ipvsadm
func (s *Service) UpdateAndInstall() error {
	if !s.instanceScope.RepositoryEnabled() {
		return nil
	}

	svc := s.getRepositoryService(s.os)
	if svc == nil {
		checkDeb, debErr := s.sshClient.SudoCmd("which apt")
		if debErr == nil && strings.Contains(checkDeb, "bin") {
			svc = repository.NewDeb(s.sshClient)
		}
		checkRpm, rpmErr := s.sshClient.SudoCmd("which yum")
		if rpmErr == nil && strings.Contains(checkRpm, "bin") {
			svc = repository.NewRPM(s.sshClient)
		}

		if debErr != nil && rpmErr != nil {
			return errors.Errorf("failed to find package manager: %v, %v", debErr, rpmErr)
		} else if debErr == nil && rpmErr == nil {
			return errors.New("can't detect the main package repository, only one of apt or yum is supported")
		}
	}

	if s.instanceScope.RepositoryUseISO() {
		if err := svc.Add(s.mountPath); err != nil {
			return errors.Wrapf(err, "failed to add local repository %s", s.mountPath)
		}
	}

	if s.instanceScope.Repository().Update {
		if err := svc.Update(); err != nil {
			return errors.Wrap(err, "failed to update os repository")
		}
	}
	if err := svc.Install(s.instanceScope.Repository().Packages...); err != nil {
		return errors.Wrap(err, "failed to use the repository to install software")
	}
	return nil
}
