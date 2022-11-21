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
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/kubesphere/kubekey/cmd/kk/pkg/bootstrap/os/repository"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/cmd/kk/pkg/utils"
	"github.com/kubesphere/kubekey/util/osrelease"
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
	etcdFiles = []string{
		"/usr/local/bin/etcd",
		"/etc/ssl/etcd",
		"/var/lib/etcd",
		"/etc/etcd.env",
	}
	clusterFiles = []string{
		"/etc/kubernetes",
		"/etc/systemd/system/etcd.service",
		"/etc/systemd/system/backup-etcd.service",
		"/etc/systemd/system/backup-etcd.timer",
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
		"ipvsadm -C",
		"ip link del kube-ipvs0",
		"ip link del nodelocaldns",
	}
)

type ResetNetworkConfig struct {
	common.KubeAction
}

func (r *ResetNetworkConfig) Execute(runtime connector.Runtime) error {
	for _, cmd := range networkResetCmds {
		_, _ = runtime.GetRunner().SudoCmd(cmd, true)
	}
	return nil
}

type UninstallETCD struct {
	common.KubeAction
}

func (s *UninstallETCD) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmd("systemctl stop etcd && exit 0", false)
	for _, file := range etcdFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), true)
	}
	return nil
}

type RemoveNodeFiles struct {
	common.KubeAction
}

func (r *RemoveNodeFiles) Execute(runtime connector.Runtime) error {
	nodeFiles := []string{
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
		"/etc/systemd/system/kubelet.service",
		"/etc/systemd/system/kubelet.service.d",
		"/usr/local/bin/kubelet",
		"/usr/local/bin/kubeadm",
		"/usr/bin/kubelet",
		"/tmp/kubekey",
		"/etc/kubekey",
	}

	for _, file := range nodeFiles {
		_, _ = runtime.GetRunner().SudoCmd(fmt.Sprintf("rm -rf %s", file), true)
	}
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

	// try to restart the cotainerd after /etc/cni has been removed
	_, _ = runtime.GetRunner().SudoCmd("systemctl restart containerd", false)
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

	host := runtime.RemoteHost()
	// type: *osrelease.data
	host.GetCache().Set(Release, osrData)
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
	src := filepath.Join(runtime.GetWorkDir(), "repository", host.GetArch(), r.ID, r.VersionID, fileName)
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

type NewRepoClient struct {
	common.KubeAction
}

func (n *NewRepoClient) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	release, ok := host.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by host cache")
	}
	r := release.(*osrelease.Data)

	repo, err := repository.New(r.ID)
	if err != nil {
		checkDeb, debErr := runtime.GetRunner().SudoCmd("which apt", false)
		if debErr == nil && strings.Contains(checkDeb, "bin") {
			repo = repository.NewDeb()
		}
		checkRPM, rpmErr := runtime.GetRunner().SudoCmd("which yum", false)
		if rpmErr == nil && strings.Contains(checkRPM, "bin") {
			repo = repository.NewRPM()
		}

		if debErr != nil && rpmErr != nil {
			return errors.Wrap(errors.WithStack(err), "new repository manager failed")
		} else if debErr == nil && rpmErr == nil {
			return errors.New("can't detect the main package repository, only one of apt or yum is supported")
		}
	}

	host.GetCache().Set("repo", repo)
	return nil
}

type BackupOriginalRepository struct {
	common.KubeAction
}

func (b *BackupOriginalRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	r, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	repo := r.(repository.Interface)

	if err := repo.Backup(runtime); err != nil {
		return errors.Wrap(errors.WithStack(err), "backup repository failed")
	}

	return nil
}

type AddLocalRepository struct {
	common.KubeAction
}

func (a *AddLocalRepository) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	r, ok := host.GetCache().Get("repo")
	if !ok {
		return errors.New("get repo failed by host cache")
	}
	repo := r.(repository.Interface)

	if installErr := repo.Add(runtime, filepath.Join(common.TmpDir, "iso")); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "add local repository failed")
	}
	if installErr := repo.Update(runtime); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "update local repository failed")
	}

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

	var pkg []string
	if _, ok := r.(*repository.Debian); ok {
		pkg = i.KubeConf.Cluster.System.Debs
	} else if _, ok := r.(*repository.RedhatPackageManager); ok {
		pkg = i.KubeConf.Cluster.System.Rpms
	}

	if installErr := r.Install(runtime, pkg...); installErr != nil {
		return errors.Wrap(errors.WithStack(installErr), "install repository package failed")
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

	var resetErr error
	defer func() {
		if resetErr != nil {
			mountPath := filepath.Join(common.TmpDir, "iso")
			umountCmd := fmt.Sprintf("umount %s", mountPath)
			_, _ = runtime.GetRunner().SudoCmd(umountCmd, false)
		}
	}()

	if resetErr = re.Reset(runtime); resetErr != nil {
		return errors.Wrap(errors.WithStack(resetErr), "reset repository failed")
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

type NodeConfigureNtpServer struct {
	common.KubeAction
}

func (n *NodeConfigureNtpServer) Execute(runtime connector.Runtime) error {

	currentHost := runtime.RemoteHost()
	release, ok := currentHost.GetCache().Get(Release)
	if !ok {
		return errors.New("get os release failed by host cache")
	}
	r := release.(*osrelease.Data)

	chronyConfigFile := "/etc/chrony.conf"
	chronyService := "chronyd.service"
	if r.ID == "ubuntu" || r.ID == "debian" {
		chronyConfigFile = "/etc/chrony/chrony.conf"
		chronyService = "chrony.service"
	}

	// if NtpServers was configured
	for _, server := range n.KubeConf.Cluster.System.NtpServers {

		serverAddr := strings.Trim(server, " \"")
		if serverAddr == currentHost.GetName() || serverAddr == currentHost.GetInternalAddress() {
			allowClientCmd := fmt.Sprintf(`sed -i '/#allow/ a\allow 0.0.0.0/0' %s`, chronyConfigFile)
			if _, err := runtime.GetRunner().SudoCmd(allowClientCmd, false); err != nil {
				return errors.Wrapf(err, "change host:%s chronyd conf failed, please check file %s", serverAddr, chronyConfigFile)
			}
		}

		// use internal ip to client chronyd server
		for _, host := range runtime.GetAllHosts() {
			if serverAddr == host.GetName() {
				serverAddr = host.GetInternalAddress()
				break
			}
		}

		checkOrAddCmd := fmt.Sprintf(`grep -q '^server %s iburst' %s||sed '1a server %s iburst' -i %s`, serverAddr, chronyConfigFile, serverAddr, chronyConfigFile)
		if _, err := runtime.GetRunner().SudoCmd(checkOrAddCmd, false); err != nil {
			return errors.Wrapf(err, "set ntpserver: %s failed, please check file %s", serverAddr, chronyConfigFile)
		}

	}

	// if Timezone was configured
	if len(n.KubeConf.Cluster.System.Timezone) > 0 {
		setTimeZoneCmd := fmt.Sprintf("timedatectl set-timezone %s", n.KubeConf.Cluster.System.Timezone)
		if _, err := runtime.GetRunner().SudoCmd(setTimeZoneCmd, false); err != nil {
			return errors.Wrapf(err, "set timezone: %s failed", n.KubeConf.Cluster.System.Timezone)
		}

		if _, err := runtime.GetRunner().SudoCmd("timedatectl set-ntp true", false); err != nil {
			return errors.Wrap(err, "timedatectl set-ntp true failed")
		}
	}

	// ensure chronyd was enabled and work normally
	if len(n.KubeConf.Cluster.System.NtpServers) > 0 || len(n.KubeConf.Cluster.System.Timezone) > 0 {
		startChronyCmd := fmt.Sprintf("systemctl enable %s && systemctl restart %s", chronyService, chronyService)
		if _, err := runtime.GetRunner().SudoCmd(startChronyCmd, false); err != nil {
			return errors.Wrap(err, "restart chronyd failed")
		}

		// tells chronyd to cancel any remaining correction that was being slewed and jump the system clock by the equivalent amount, making it correct immediately.
		if _, err := runtime.GetRunner().SudoCmd("chronyc makestep > /dev/null && chronyc sources", true); err != nil {
			return errors.Wrap(err, "chronyc makestep failed")
		}
	}

	return nil
}
