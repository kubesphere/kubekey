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

package etcd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	kubekeyapiv1alpha2 "github.com/kubesphere/kubekey/v3/cmd/kk/apis/kubekey/v1alpha2"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/common"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/action"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/util"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/etcd/templates"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/files"
	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/utils"
)

type EtcdNode struct {
	NodeName  string
	EtcdName  string
	EtcdExist bool
}

type EtcdCluster struct {
	clusterExist    bool
	accessAddresses string
	peerAddresses   []string
}

const (
	LocalCertsDir = "localCertsDir"
	CertsFileList = "certsFileList"

	NewCluster   = "new"
	ExistCluster = "existing"
)

type GetStatus struct {
	common.KubeAction
}

func (g *GetStatus) Execute(runtime connector.Runtime) error {
	exist, err := runtime.GetRunner().FileExist("/etc/etcd.env")
	if err != nil {
		return err
	}

	host := runtime.RemoteHost()
	cluster := &EtcdCluster{
		clusterExist:    true,
		accessAddresses: "",
		peerAddresses:   []string{},
	}

	if exist {
		etcdEnv, err := runtime.GetRunner().SudoCmd("cat /etc/etcd.env | grep ETCD_NAME", true)
		if err != nil {
			return err
		}

		etcdName := etcdEnv[strings.Index(etcdEnv, "=")+1:]
		// type: string
		host.GetCache().Set(common.ETCDName, etcdName)
		// type: bool
		host.GetCache().Set(common.ETCDExist, true)

		if v, ok := g.PipelineCache.Get(common.ETCDCluster); ok {
			c := v.(*EtcdCluster)
			c.peerAddresses = append(c.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetInternalAddress()))
			c.clusterExist = true
			// type: *EtcdCluster
			g.PipelineCache.Set(common.ETCDCluster, c)
		} else {
			cluster.peerAddresses = append(cluster.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetInternalAddress()))
			cluster.clusterExist = true
			g.PipelineCache.Set(common.ETCDCluster, cluster)
		}
	} else {
		host.GetCache().Set(common.ETCDName, fmt.Sprintf("etcd-%s", host.GetName()))
		host.GetCache().Set(common.ETCDExist, false)

		if _, ok := g.PipelineCache.Get(common.ETCDCluster); !ok {
			cluster.clusterExist = false
			g.PipelineCache.Set(common.ETCDCluster, cluster)
		}
	}
	return nil
}

type SyncCertsFile struct {
	common.KubeAction
}

func (s *SyncCertsFile) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.ModuleCache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get etcd local certs dir by module cache failed")
	}
	files, ok := s.ModuleCache.Get(CertsFileList)
	if !ok {
		return errors.New("get etcd certs file list by module cache failed")
	}
	dir := localCertsDir.(string)
	fileList := files.([]string)

	for _, fileName := range fileList {
		if err := runtime.GetRunner().SudoScp(filepath.Join(dir, fileName), filepath.Join(common.ETCDCertDir, fileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "scp etcd certs file failed")
		}
	}

	return nil
}

type InstallETCDBinary struct {
	common.KubeAction
}

func (g *InstallETCDBinary) Execute(runtime connector.Runtime) error {
	if err := utils.ResetTmpDir(runtime); err != nil {
		return err
	}

	binariesMapObj, ok := g.PipelineCache.Get(common.KubeBinaries + "-" + runtime.RemoteHost().GetArch())
	if !ok {
		return errors.New("get KubeBinary by pipeline cache failed")
	}
	binariesMap := binariesMapObj.(map[string]*files.KubeBinary)
	binary, ok := binariesMap["etcd"]
	if !ok {
		return fmt.Errorf("get kube binary etcd info failed: no such key")
	}

	dst := filepath.Join(common.TmpDir, binary.FileName)
	if err := runtime.GetRunner().Scp(binary.Path(), dst); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync etcd tar.gz failed")
	}

	etcdDir := strings.TrimSuffix(binary.FileName, ".tar.gz")
	installCmd := fmt.Sprintf("tar -zxf %s && cp -f %s/etcd* /usr/local/bin/ && chmod +x /usr/local/bin/etcd* && rm -rf %s", dst, etcdDir, etcdDir)
	if _, err := runtime.GetRunner().SudoCmd(installCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "install etcd binaries failed")
	}
	return nil
}

type GenerateAccessAddress struct {
	common.KubeAction
}

func (g *GenerateAccessAddress) Execute(runtime connector.Runtime) error {
	var addrList []string
	for _, host := range runtime.GetHostsByRole(common.ETCD) {
		addrList = append(addrList, fmt.Sprintf("https://%s:2379", host.GetInternalAddress()))
	}

	accessAddresses := strings.Join(addrList, ",")
	if v, ok := g.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		cluster.accessAddresses = accessAddresses
		g.PipelineCache.Set(common.ETCDCluster, cluster)
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

type HealthCheck struct {
	common.KubeAction
}

func (h *HealthCheck) Execute(runtime connector.Runtime) error {
	if v, ok := h.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		if err := healthCheck(runtime, cluster); err != nil {
			return err
		}
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

func healthCheck(runtime connector.Runtime, cluster *EtcdCluster) error {
	host := runtime.RemoteHost()
	checkHealthCmd := fmt.Sprintf("export ETCDCTL_API=2;"+
		"export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';"+
		"export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';"+
		"export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';"+
		"%s/etcdctl --endpoints=%s cluster-health | grep -q 'cluster is healthy'",
		host.GetName(), host.GetName(), common.BinDir, cluster.accessAddresses)
	if _, err := runtime.GetRunner().SudoCmd(checkHealthCmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "etcd health check failed")
	}
	return nil
}

type GenerateConfig struct {
	common.KubeAction
}

func (g *GenerateConfig) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	etcdName, ok := host.GetCache().GetMustString(common.ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := g.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)

		cluster.peerAddresses = append(cluster.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetInternalAddress()))
		g.PipelineCache.Set(common.ETCDCluster, cluster)

		if !cluster.clusterExist {
			if err := refreshConfig(runtime, cluster.peerAddresses, NewCluster, etcdName); err != nil {
				return err
			}
		} else {
			if err := refreshConfig(runtime, cluster.peerAddresses, ExistCluster, etcdName); err != nil {
				return err
			}
		}
		return nil
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
}

type RefreshConfig struct {
	common.KubeAction
	ToExisting bool
}

func (r *RefreshConfig) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	etcdName, ok := host.GetCache().GetMustString(common.ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := r.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)

		if r.ToExisting {
			if err := refreshConfig(runtime, cluster.peerAddresses, ExistCluster, etcdName); err != nil {
				return err
			}
			return nil
		}

		if !cluster.clusterExist {
			if err := refreshConfig(runtime, cluster.peerAddresses, NewCluster, etcdName); err != nil {
				return err
			}
		} else {
			if err := refreshConfig(runtime, cluster.peerAddresses, ExistCluster, etcdName); err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("get etcd cluster status by pipeline cache failed")
}

func refreshConfig(runtime connector.Runtime, endpoints []string, state, etcdName string) error {
	host := runtime.RemoteHost()

	UnsupportedArch := false
	if host.GetArch() != "amd64" {
		UnsupportedArch = true
	}

	templateAction := action.Template{
		Template: templates.EtcdEnv,
		Dst:      filepath.Join("/etc/", templates.EtcdEnv.Name()),
		Data: util.Data{
			"Tag":             kubekeyapiv1alpha2.DefaultEtcdVersion,
			"Name":            etcdName,
			"Ip":              host.GetInternalAddress(),
			"Hostname":        host.GetName(),
			"State":           state,
			"peerAddresses":   strings.Join(endpoints, ","),
			"UnsupportedArch": UnsupportedArch,
			"Arch":            host.GetArch(),
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}
	return nil
}

type JoinMember struct {
	common.KubeAction
}

func (j *JoinMember) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	etcdName, ok := host.GetCache().GetMustString(common.ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := j.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		joinMemberCmd := fmt.Sprintf("export ETCDCTL_API=2;"+
			"export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';"+
			"export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';"+
			"export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';"+
			"%s/etcdctl --endpoints=%s member add %s %s",
			host.GetName(), host.GetName(), common.BinDir, cluster.accessAddresses, etcdName,
			fmt.Sprintf("https://%s:2380", host.GetInternalAddress()))

		if _, err := runtime.GetRunner().SudoCmd(joinMemberCmd, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "add etcd member failed")
		}
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

type CheckMember struct {
	common.KubeAction
}

func (c *CheckMember) Execute(runtime connector.Runtime) error {
	host := runtime.RemoteHost()
	if v, ok := c.PipelineCache.Get(common.ETCDCluster); ok {
		cluster := v.(*EtcdCluster)
		checkMemberCmd := fmt.Sprintf("export ETCDCTL_API=2;"+
			"export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';"+
			"export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';"+
			"export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';"+
			"%s/etcdctl --no-sync --endpoints=%s member list", host.GetName(), host.GetName(), common.BinDir, cluster.accessAddresses)
		memberList, err := runtime.GetRunner().SudoCmd(checkMemberCmd, true)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "list etcd member failed")
		}
		if !strings.Contains(memberList, fmt.Sprintf("https://%s:2379", host.GetInternalAddress())) {
			return errors.Wrap(errors.WithStack(err), "add etcd member failed")
		}
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

type RestartETCD struct {
	common.KubeAction
}

func (r *RestartETCD) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload && systemctl restart etcd && systemctl enable etcd", true); err != nil {
		return errors.Wrap(errors.WithStack(err), "start etcd failed")
	}
	return nil
}

type BackupETCD struct {
	common.KubeAction
}

func (b *BackupETCD) Execute(runtime connector.Runtime) error {
	templateAction := action.Template{
		Template: templates.EtcdBackupScript,
		Dst:      filepath.Join(b.KubeConf.Cluster.Etcd.BackupScriptDir, "etcd-backup.sh"),
		Data: util.Data{
			"Hostname":            runtime.RemoteHost().GetName(),
			"Etcdendpoint":        fmt.Sprintf("https://%s:2379", runtime.RemoteHost().GetInternalAddress()),
			"Backupdir":           b.KubeConf.Cluster.Etcd.BackupDir,
			"KeepbackupNumber":    b.KubeConf.Cluster.Etcd.KeepBackupNumber + 1,
			"EtcdBackupScriptDir": b.KubeConf.Cluster.Etcd.BackupScriptDir,
		},
	}

	templateAction.Init(nil, nil)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/etcd-backup.sh", b.KubeConf.Cluster.Etcd.BackupScriptDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod etcd backup script failed")
	}
	return nil
}

type EnableBackupETCDService struct {
	common.KubeAction
}

func (e *EnableBackupETCDService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl enable --now backup-etcd.timer",
		false); err != nil {
		return errors.Wrap(errors.WithStack(err), "enable backup-etcd.service failed")
	}
	return nil
}
