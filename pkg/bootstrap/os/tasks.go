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

package os

import (
	"fmt"
	osrelease "github.com/dominodatalab/os-release"
	"github.com/kubesphere/kubekey/pkg/bootstrap/os/repository"
	"github.com/kubesphere/kubekey/pkg/common"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/utils"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

type NodeConfigureOS struct {
	common.KubeAction
}

func (n *NodeConfigureOS) Execute(runtime connector.Runtime) error {

	host := runtime.RemoteHost()
	if err := addUsers(runtime, host); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to add users")
	}

	if err := createDirectories(runtime, host); err != nil {
		return err
	}

	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	_, err1 := runtime.GetRunner().SudoCmd(fmt.Sprintf("hostnamectl set-hostname %s && sed -i '/^127.0.1.1/s/.*/127.0.1.1      %s/g' /etc/hosts", host.GetName(), host.GetName()), false)
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to override hostname")
	}

	return nil
}

func addUsers(runtime connector.Runtime, node connector.Host) error {
	if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Kubernetes user' -s /sbin/nologin -r kube || :", false); err != nil {
		return err
	}

	if node.IsRole(common.ETCD) {
		if _, err := runtime.GetRunner().SudoCmd("useradd -M -c 'Etcd user' -s /sbin/nologin -r etcd || :", false); err != nil {
			return err
		}
	}

	return nil
}

func createDirectories(runtime connector.Runtime, node connector.Host) error {
	dirs := []string{
		common.BinDir,
		common.KubeConfigDir,
		common.KubeCertDir,
		common.KubeManifestDir,
		common.KubeScriptDir,
		common.KubeletFlexvolumesPluginsDir,
	}

	for _, dir := range dirs {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s", dir), false); err != nil {
			return err
		}
		if dir == common.KubeletFlexvolumesPluginsDir {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", "/usr/libexec/kubernetes"), false); err != nil {
				return err
			}
		} else {
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chown kube -R %s", dir), false); err != nil {
				return err
			}
		}
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/etc/cni/net.d", "/etc/cni"), false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/opt/cni/bin", "/opt/cni"), false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown kube -R %s", "/var/lib/calico", "/var/lib/calico"), false); err != nil {
		return err
	}

	if node.IsRole(common.ETCD) {
		if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("mkdir -p %s && chown etcd -R %s", "/var/lib/etcd", "/var/lib/etcd"), false); err != nil {
			return err
		}
	}

	return nil
}

type NodeExecScript struct {
	common.KubeAction
}

func (n *NodeExecScript) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/initOS.sh", common.KubeScriptDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to chmod +x init os script")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s/initOS.sh", common.KubeScriptDir), true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to configure operating system")
	}
	return nil
}

var (
	clusterFiles = []string{
		"/usr/local/bin/etcd",
		"/etc/ssl/etcd",
		"/var/lib/etcd",
		"/etc/etcd.env",
		"/etc/kubernetes",
		"/etc/systemd/system/etcd.service",
		"/var/log/calico",
		"/etc/cni",
		"/var/log/pods/",
		"/var/lib/cni",
		"/var/lib/calico",
		"/var/lib/kubelet",
		"/run/calico",
		"/run/flannel",
		"/etc/flannel",
		"/var/openebs",
		"/etc/systemd/system/kubelet.service",
		"/etc/systemd/system/kubelet.service.d",
		"/usr/local/bin/kubelet",
		"/usr/local/bin/kubeadm",
		"/usr/bin/kubelet",
		"/var/lib/rook",
		"/tmp/kubekey",
		"/etc/kubekey",
	}

	networkResetCmds = []string{
		"iptables -F",
		"iptables -X",
		"iptables -F -t nat",
		"iptables -X -t nat",
		"ip link del kube-ipvs0",
		"ip link del nodelocaldns",
	}
)

type ResetNetworkConfig struct {
	common.KubeAction
}

func (r *ResetNetworkConfig) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd(strings.Join(networkResetCmds, " && "), true)
	return nil
}

type StopETCDService struct {
	common.KubeAction
}

func (s *StopETCDService) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop etcd && exit 0", false)
	return nil
}

type RemoveFiles struct {
	common.KubeAction
}

func (r *RemoveFiles) Execute(runtime connector.Runtime) error {
	for _, file := range clusterFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), true)
	}
	return nil
}

type DaemonReload struct {
	common.KubeAction
}

func (d *DaemonReload) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && exit 0", false); err != nil {
		return errors.Wrap(errors.WithStack(err), "systemctl daemon-reload failed")
	}
	return nil
}

type GetOSData struct {
	common.KubeAction
}

func (g *GetOSData) Execute(runtime connector.Runtime) error {
	osReleaseStr, err := runtime.GetRunner().SudoCmd("cat /etc/os-release", false)
	if err != nil {
		return err
	}
	osrData := osrelease.Parse(strings.Replace(osReleaseStr, "\r\n", "\n", -1))

	pkgToolStr, err := runtime.GetRunner().SudoCmd(
		"if [ ! -z $(which yum 2>/dev/null) ]; "+
			"then echo rpm; "+
			"elif [ ! -z $(which apt 2>/dev/null) ]; "+
			"then echo deb; "+
			"fi", false)
	if err != nil {
		return err
	}

	host := runtime.RemoteHost()
	// type: *osrelease.data
	host.GetCache().Set(Release, osrData)
	// type: string
	host.GetCache().Set(PkgTool, pkgToolStr)
	return nil
}

type OnlineInstallDependencies struct {
	common.KubeAction
}

func (o *OnlineInstallDependencies) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	pkg, ok := host.GetCache().GetMustString(PkgTool)
	if !ok {
		return errors.New("get pkgTool failed by root cache")
	}
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by root cache")
	}
	r := release.(*osrelease.Data)

	switch strings.TrimSpace(pkg) {
	case "deb":
		if _, err := runtime.GetRunner().SudoCmd(
			"apt update;"+
				"apt install socat conntrack ipset ebtables chrony -y;",
			false); err != nil {
			return err
		}
	case "rpm":
		if _, err := runtime.GetRunner().SudoCmd(
			"yum install yum-utils openssl socat conntrack ipset ebtables chrony -y",
			false); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported operating system: %s", r.ID))
	}
	return nil
}

type OfflineInstallDependencies struct {
	common.KubeAction
}

func (o *OfflineInstallDependencies) Execute(runtime connector.Runtime) error {
	fp, err := filepath.Abs(o.KubeConf.Arg.SourcesDir)
	if err != nil {
		return errors.Wrap(err, "Failed to look up current directory")
	}

	host := runtime.RemoteHost()
	pkg, ok := host.GetCache().GetMustString(PkgTool)
	if !ok {
		return errors.New("get pkgTool failed by root cache")
	}
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by root cache")
	}
	r := release.(*osrelease.Data)

	switch strings.TrimSpace(pkg) {
	case "deb":
		dirName := fmt.Sprintf("%s-%s-%s-debs", r.ID, r.VersionID, host.GetArch())
		srcPath := filepath.Join(fp, dirName)
		srcTar := fmt.Sprintf("%s.tar.gz", srcPath)
		dstPath := filepath.Join(common.TmpDir, dirName)
		dstTar := fmt.Sprintf("%s.tar.gz", dstPath)

		_ = runtime.GetRunner().Scp(srcTar, dstTar)
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("tar -zxvf %s -C %s && dpkg -iR --force-all %s", dstTar, common.TmpDir, dstPath),
			false); err != nil {
			return err
		}
	case "rpm":
		dirName := fmt.Sprintf("%s-%s-%s-rpms", r.ID, r.VersionID, host.GetArch())
		srcPath := filepath.Join(fp, dirName)
		srcTar := fmt.Sprintf("%s.tar.gz", srcPath)
		dstPath := filepath.Join(common.TmpDir, dirName)
		dstTar := fmt.Sprintf("%s.tar.gz", dstPath)

		_ = runtime.GetRunner().Scp(srcTar, dstTar)
		if _, err := runtime.GetRunner().SudoCmd(
			fmt.Sprintf("tar -zxvf %s -C %s && rpm -Uvh --force --nodeps %s", dstTar, common.TmpDir, filepath.Join(dstPath, "*rpm")),
			false); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported operating system: %s", r.ID))
	}
	return nil
}

type SyncRepositoryFile struct {
	common.KubeAction
}

func (s *SyncRepositoryFile) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return errors.Wrap(err, "reset tmp dir failed")
	}

	host := runtime.RemoteHost()
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by root cache")
	}
	r := release.(*osrelease.Data)

	fileName := fmt.Sprintf("%s-%s-%s.iso", r.ID, r.VersionID, host.GetArch())
	src := filepath.Join(runtime.GetWorkDir(), common.Artifact, "repository", host.GetArch(), r.ID, r.VersionID, fileName)
	dst := filepath.Join(common.TmpDir, fileName)
	if err := runtime.GetRunner().Scp(src, dst); err != nil {
		return errors.Wrapf(errors.WithStack(err), "scp %s to %s failed", src, dst)
	}

	host.GetCache().Set("iso", fileName)
	return nil
}

type MountISO struct {
	common.KubeAction
}

func (m *MountISO) Execute(runtime connector.Runtime) error {
	mountPath := filepath.Join(common.TmpDir, "iso")
	if err := runtime.GetRunner().MkDir(mountPath); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create mount dir failed")
	}

	host := runtime.RemoteHost()
	isoFile, _ := host.GetCache().GetMustString("iso")
	path := filepath.Join(common.TmpDir, isoFile)
	mountCmd := fmt.Sprintf("sudo mount -t iso9660 -o loop %s %s", path, mountPath)
	if _, err := runtime.GetRunner().Cmd(mountCmd, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "mount %s at %s failed", path, mountPath)
	}
	return nil
}

type AddLocalRepository struct {
	common.KubeAction
}

func (a *AddLocalRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by host cache")
	}
	r := release.(*osrelease.Data)

	repo, err := repository.New(r.ID, runtime)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "new repository manager failed")
	}

	if err := repo.Backup(); err != nil {
		return errors.Wrap(errors.WithStack(err), "backup repository failed")
	}
	if err := repo.Add(filepath.Join(common.TmpDir, "iso")); err != nil {
		return errors.Wrap(errors.WithStack(err), "add local repository failed")
	}
	if err := repo.Update(); err != nil {
		return errors.Wrap(errors.WithStack(err), "update local repository failed")
	}

	host.GetCache().Set("repo", repo)
	return nil
}

type InstallPackage struct {
	common.KubeAction
}

func (i *InstallPackage) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	repo, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	r := repo.(repository.Interface)

	if err := r.Install(); err != nil {
		return errors.Wrap(errors.WithStack(err), "install repository package failed")
	}
	return nil
}

type ResetRepository struct {
	common.KubeAction
}

func (r *ResetRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	repo, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	re := repo.(repository.Interface)

	if err := re.Reset(); err != nil {
		return errors.Wrap(errors.WithStack(err), "reset repository failed")
	}

	return nil
}

type UmountISO struct {
	common.KubeAction
}

func (u *UmountISO) Execute(runtime connector.Runtime) error {
	mountPath := filepath.Join(common.TmpDir, "iso")
	umountCmd := fmt.Sprintf("umount %s", mountPath)
	if _, err := runtime.GetRunner().SudoCmd(umountCmd, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "umount %s failed", mountPath)
	}
	return nil
}
