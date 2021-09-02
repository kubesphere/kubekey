package etcd

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/core/connector"
	"github.com/kubesphere/kubekey/pkg/core/util"
	"github.com/kubesphere/kubekey/pkg/pipelines/common"
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
	ETCDCluster   = "etcdCluster"
	LocalCertsDir = "localCertsDir"
	CertsFileList = "certsFileList"
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

		n := EtcdNode{
			NodeName:  host.GetName(),
			EtcdName:  etcdEnv[strings.Index(etcdEnv, "=")+1:],
			EtcdExist: true,
		}
		g.Cache.Set(host.GetName(), n)

		if v, ok := g.Cache.Get(ETCDCluster); ok {
			c := v.(EtcdCluster)
			c.peerAddresses = append(c.peerAddresses, fmt.Sprintf("%s=https://%s:2380", n.EtcdName, host.GetAddress()))
			c.clusterExist = true
			g.Cache.Set(ETCDCluster, c)
		} else {
			cluster.peerAddresses = append(cluster.peerAddresses, fmt.Sprintf("%s=https://%s:2380", n.EtcdName, host.GetAddress()))
			cluster.clusterExist = true
			g.Cache.Set(ETCDCluster, cluster)
		}
	} else {
		n := EtcdNode{
			NodeName:  host.GetName(),
			EtcdName:  fmt.Sprintf("etcd-%s", host.GetName()),
			EtcdExist: false,
		}
		g.Cache.Set(host.GetName(), n)

		if _, ok := g.Cache.Get(ETCDCluster); !ok {
			cluster.clusterExist = false
			g.Cache.Set(ETCDCluster, cluster)
		}
	}
	return nil
}

type FirstETCDNode struct {
	common.KubePrepare
	Not bool
}

func (f *FirstETCDNode) PreCheck(runtime connector.Runtime) (bool, error) {
	v, ok := f.Cache.Get(ETCDCluster)
	if !ok {
		return false, errors.New("get etcd cluster status by module cache failed")
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

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("cp -r %s %s", common.ETCDCertDir, common.TmpDir), false); err != nil {
		return errors.Wrap(errors.WithStack(err), "copy certs result failed")
	}

	localCertsDir := filepath.Join(runtime.GetWorkDir(), "ETCD_certs")
	if err := util.CreateDir(localCertsDir); err != nil {
		return err
	}

	files := generateCertsFiles(runtime)
	for _, fileName := range files {
		if err := runtime.GetRunner().Fetch(filepath.Join(localCertsDir, fileName), filepath.Join(common.TmpDir, fileName)); err != nil {
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
	if v, ok := g.Cache.Get(ETCDCluster); ok {
		cluster := v.(EtcdCluster)
		cluster.accessAddresses = accessAddresses
		g.Cache.Set(ETCDCluster, cluster)
	} else {
		return errors.New("get etcd cluster status by module cache failed")
	}
	return nil
}
