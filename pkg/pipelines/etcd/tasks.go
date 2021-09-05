package etcd

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/action"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
	"github.com/kubesphere/kubekey/pkg/pipelines/etcd/templates"
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
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
	ETCDName  = "etcdName"
	ETCDExist = "etcdExist"

	ETCDCluster   = "etcdCluster"
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

	host := g.Runtime.RemoteHost()
	cluster := EtcdCluster{
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
		host.SetLabel(ETCDName, etcdName)
		host.SetLabel(ETCDExist, "true")

		if v, ok := g.RootCache.Get(ETCDCluster); ok {
			c := v.(EtcdCluster)
			c.peerAddresses = append(c.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetAddress()))
			c.clusterExist = true
			g.RootCache.Set(ETCDCluster, c)
		} else {
			cluster.peerAddresses = append(cluster.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetAddress()))
			cluster.clusterExist = true
			g.RootCache.Set(ETCDCluster, cluster)
		}
	} else {
		host.SetLabel(ETCDName, fmt.Sprintf("etcd-%s", host.GetName()))
		host.SetLabel(ETCDExist, "false")

		if _, ok := g.RootCache.Get(ETCDCluster); !ok {
			cluster.clusterExist = false
			g.RootCache.Set(ETCDCluster, cluster)
		}
	}
	return nil
}

type FirstETCDNode struct {
	common.KubePrepare
	Not bool
}

func (f *FirstETCDNode) PreCheck(runtime connector.Runtime) (bool, error) {
	v, ok := f.RootCache.Get(ETCDCluster)
	if !ok {
		return false, errors.New("get etcd cluster status by pipeline cache failed")
	}
	cluster := v.(EtcdCluster)

	if (!cluster.clusterExist && runtime.GetHostsByRole(common.ETCD)[0].GetName() == runtime.RemoteHost().GetName()) ||
		(cluster.clusterExist && strings.Contains(cluster.peerAddresses[0], runtime.RemoteHost().GetInternalAddress())) {
		return !f.Not, nil
	}
	return f.Not, nil
}

type ExecCertsScript struct {
	common.KubeAction
}

func (e *ExecCertsScript) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/make-ssl-etcd.sh", common.ETCDCertDir), false)
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("sh %s/make-ssl-etcd.sh -f %s/openssl.conf -d %s", common.ETCDCertDir, common.ETCDCertDir, common.ETCDCertDir)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "generate etcd certs faild")
	}

	tmpCertsDir := filepath.Join(common.TmpDir, "ETCD_certs")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cp -r %s %s", common.ETCDCertDir, tmpCertsDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy certs result failed")
	}

	localCertsDir := filepath.Join(runtime.GetWorkDir(), "ETCD_certs")
	if err := util.CreateDir(localCertsDir); err != nil {
		return err
	}

	files := generateCertsFiles(runtime)
	for _, fileName := range files {
		if err := runtime.GetRunner().Fetch(filepath.Join(localCertsDir, fileName), filepath.Join(tmpCertsDir, fileName)); err != nil {
			return errors.Wrap(errors.WithStack(err), "fetch etcd certs file failed")
		}
	}

	e.Cache.Set(LocalCertsDir, localCertsDir)
	e.Cache.Set(CertsFileList, files)
	return nil
}

func generateCertsFiles(runtime connector.Runtime) []string {
	var certsList []string
	certsList = append(certsList, "ca.pem")
	certsList = append(certsList, "ca-key.pem")
	for _, host := range runtime.GetHostsByRole(common.ETCD) {
		certsList = append(certsList, fmt.Sprintf("admin-%s.pem", host.GetName()))
		certsList = append(certsList, fmt.Sprintf("admin-%s-key.pem", host.GetName()))
		certsList = append(certsList, fmt.Sprintf("member-%s.pem", host.GetName()))
		certsList = append(certsList, fmt.Sprintf("member-%s-key.pem", host.GetName()))
	}
	for _, host := range runtime.GetHostsByRole(common.ETCD) {
		certsList = append(certsList, fmt.Sprintf("node-%s.pem", host.GetName()))
		certsList = append(certsList, fmt.Sprintf("node-%s-key.pem", host.GetName()))
	}
	return certsList
}

type SyncCertsFile struct {
	common.KubeAction
}

func (s *SyncCertsFile) Execute(runtime connector.Runtime) error {
	localCertsDir, ok := s.Cache.Get(LocalCertsDir)
	if !ok {
		return errors.New("get etcd local certs dir by module cache failed")
	}
	files, ok := s.Cache.Get(CertsFileList)
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
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("if [ -d %s ]; then rm -rf %s ;fi && mkdir -p %s", common.TmpDir, common.TmpDir, common.TmpDir), false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "reset tmp dir failed")
	}

	etcdFile := fmt.Sprintf("etcd-%s-linux-%s", kubekeyapiv1alpha1.DefaultEtcdVersion, runtime.RemoteHost().GetArch())
	filesDir := filepath.Join(runtime.GetWorkDir(), g.KubeConf.Cluster.Kubernetes.Version, runtime.RemoteHost().GetArch())
	if err := runtime.GetRunner().Scp(fmt.Sprintf("%s/%s.tar.gz", filesDir, etcdFile), fmt.Sprintf("%s/%s.tar.gz", common.TmpDir, etcdFile)); err != nil {
		return errors.Wrap(errors.WithStack(err), "sync etcd tar.gz failed")
	}

	installCmd := fmt.Sprintf("tar -zxf %s/%s.tar.gz && cp -f %s/etcd* /usr/local/bin/ && chmod +x /usr/local/bin/etcd* && rm -rf %s", common.TmpDir, etcdFile, etcdFile, etcdFile)
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
	if v, ok := g.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
		cluster.accessAddresses = accessAddresses
		g.RootCache.Set(ETCDCluster, cluster)
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

type NodeETCDExist struct {
	common.KubePrepare
	Not bool
}

func (n *NodeETCDExist) PreCheck(runtime connector.Runtime) (bool, error) {
	host := runtime.RemoteHost()
	if v, ok := host.GetLabel(ETCDExist); ok {
		if v == "true" {
			return !n.Not, nil
		} else {
			return n.Not, nil
		}
	} else {
		return false, errors.New("get etcd node status by host label failed")
	}
}

type HealthCheck struct {
	common.KubeAction
}

func (h *HealthCheck) Execute(runtime connector.Runtime) error {
	if v, ok := h.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
		if err := healthCheck(runtime, cluster); err != nil {
			return err
		}
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
	return nil
}

func healthCheck(runtime connector.Runtime, cluster EtcdCluster) error {
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
	etcdName, ok := host.GetLabel(ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := g.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)

		cluster.peerAddresses = append(cluster.peerAddresses, fmt.Sprintf("%s=https://%s:2380", etcdName, host.GetInternalAddress()))
		g.RootCache.Set(ETCDCluster, cluster)

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
	etcdName, ok := host.GetLabel(ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := r.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)

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
	} else {
		return errors.New("get etcd cluster status by pipeline cache failed")
	}
}

func refreshConfig(runtime connector.Runtime, endpoints []string, state, etcdName string) error {
	host := runtime.RemoteHost()

	UnsupportedArch := false
	if host.GetArch() != "amd64" {
		UnsupportedArch = true
	}

	templateAction := action.Template{
		Template: templates.EtcdEnv,
		Dst:      "/etc/etcd.env",
		Data: util.Data{
			"Tag":             kubekeyapiv1alpha1.DefaultEtcdVersion,
			"Name":            etcdName,
			"Ip":              host.GetInternalAddress(),
			"Hostname":        host.GetName(),
			"State":           state,
			"peerAddresses":   strings.Join(endpoints, ","),
			"UnsupportedArch": UnsupportedArch,
			"Arch":            host.GetArch(),
		},
	}

	templateAction.Init(nil, nil, runtime)
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
	etcdName, ok := host.GetLabel(ETCDName)
	if !ok {
		return errors.New("get etcd node status by host label failed")
	}

	if v, ok := j.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
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
	if v, ok := c.RootCache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
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
		Dst:      filepath.Join(b.KubeConf.Cluster.Kubernetes.EtcdBackupScriptDir, "etcd-backup.sh"),
		Data: util.Data{
			"Hostname":            runtime.RemoteHost().GetName(),
			"Etcdendpoint":        fmt.Sprintf("https://%s:2379", runtime.RemoteHost().GetInternalAddress()),
			"Backupdir":           b.KubeConf.Cluster.Kubernetes.EtcdBackupDir,
			"KeepbackupNumber":    b.KubeConf.Cluster.Kubernetes.KeepBackupNumber,
			"EtcdBackupPeriod":    b.KubeConf.Cluster.Kubernetes.EtcdBackupPeriod,
			"EtcdBackupScriptDir": b.KubeConf.Cluster.Kubernetes.EtcdBackupScriptDir,
			"EtcdBackupHour":      templates.BackupTimeInterval(runtime, b.KubeConf),
		},
	}

	templateAction.Init(nil, nil, runtime)
	if err := templateAction.Execute(runtime); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("chmod +x %s/etcd-backup.sh", b.KubeConf.Cluster.Kubernetes.EtcdBackupScriptDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "chmod etcd backup script failed")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("sh %s/etcd-backup.sh", b.KubeConf.Cluster.Kubernetes.EtcdBackupScriptDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to run the etcd-backup.sh")
	}
	return nil
}
