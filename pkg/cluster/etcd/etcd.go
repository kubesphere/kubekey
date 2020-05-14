package etcd

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/etcd/tmpl"
	"github.com/kubesphere/kubekey/pkg/images"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"time"
)

var (
	certsStr         = make(chan map[string]string)
	certsContent     = map[string]string{}
	etcdBackupPrefix = "/var/backups"
	etcdDataDir      = "/var/lib/etcd"
	etcdConfigDir    = "/etc/ssl/etcd"
	etcdCertDir      = "/etc/ssl/etcd/ssl"
	etcdBinDir       = "/usr/local/bin"
)

func GenerateEtcdCerts(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Generate etcd certs")

	return mgr.RunTaskOnEtcdNodes(generateCerts, true)
}

func generateCerts(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {

	if mgr.Runner.Index == 0 {
		certsScript, err := tmpl.GenerateEtcdSslScript(mgr)
		if err != nil {
			return err
		}
		certsScriptBase64 := base64.StdEncoding.EncodeToString([]byte(certsScript))
		_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > /tmp/kubekey/make-ssl-etcd.sh && chmod +x /tmp/kubekey/make-ssl-etcd.sh", certsScriptBase64))
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "failed to generate etcd certs script")
		}
		certsOpensslCfg, err := tmpl.GenerateEtcdSslCfg(mgr.Cluster)
		if err != nil {
			return err
		}
		certsOpensslCfgBase64 := base64.StdEncoding.EncodeToString([]byte(certsOpensslCfg))
		_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("echo %s | base64 -d > /tmp/kubekey/openssl.conf", certsOpensslCfgBase64))
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), "failed to generate etcd certs script")
		}

		cmd := fmt.Sprintf("mkdir -p %s && /bin/bash -x %s/make-ssl-etcd.sh -f %s/openssl.conf -d %s", etcdCertDir, "/tmp/kubekey", "/tmp/kubekey", etcdCertDir)

		_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd))
		if err3 != nil {
			return errors.Wrap(errors.WithStack(err3), "failed to generate etcd certs")
		}

		for _, cert := range generateCertsFiles(mgr) {
			certsBase64Cmd := fmt.Sprintf("cat %s/%s | base64 --wrap=0", etcdCertDir, cert)
			certsBase64, err4 := mgr.Runner.RunCmd(certsBase64Cmd)
			if err4 != nil {
				return errors.Wrap(errors.WithStack(err4), "failed to get etcd certs content")
			}
			certsContent[cert] = certsBase64
		}

		for i := 1; i <= len(mgr.EtcdNodes)-1; i++ {
			certsStr <- certsContent
		}

	} else {
		mgr.Runner.RunCmd(fmt.Sprintf("sudo mkdir -p %s", etcdCertDir))
		for file, cert := range <-certsStr {
			writeCertCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s/%s\"", cert, etcdCertDir, file)
			_, err4 := mgr.Runner.RunCmd(writeCertCmd)
			if err4 != nil {
				return errors.Wrap(errors.WithStack(err4), "failed to write etcd certs content")
			}
		}
	}

	return nil
}

func generateCertsFiles(mgr *manager.Manager) []string {
	var certsList []string
	certsList = append(certsList, "ca.pem")
	certsList = append(certsList, "ca-key.pem")
	for _, host := range mgr.EtcdNodes {
		certsList = append(certsList, fmt.Sprintf("admin-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("admin-%s-key.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("member-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("member-%s-key.pem", host.Name))
	}
	for _, host := range mgr.MasterNodes {
		certsList = append(certsList, fmt.Sprintf("node-%s.pem", host.Name))
		certsList = append(certsList, fmt.Sprintf("node-%s-key.pem", host.Name))
	}
	return certsList
}

func SyncEtcdCertsToMaster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Sync etcd certs")

	return mgr.RunTaskOnMasterNodes(syncEtcdCertsToMaster, true)
}

func syncEtcdCertsToMaster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if !node.IsEtcd {
		mgr.Runner.RunCmd(fmt.Sprintf("sudo mkdir -p %s", etcdCertDir))
		for file, cert := range certsContent {
			writeCertCmd := fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > %s/%s\"", cert, etcdCertDir, file)
			_, err := mgr.Runner.RunCmd(writeCertCmd)
			if err != nil {
				return errors.Wrap(errors.WithStack(err), "failed to sync etcd certs to master")
			}
		}
	}
	return nil
}

func GenerateEtcdService(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Start etcd cluster")

	return mgr.RunTaskOnEtcdNodes(generateEtcdService, true)
}

func generateEtcdService(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	etcdService, err := tmpl.GenerateEtcdService(mgr, mgr.Runner.Index)
	if err != nil {
		return err
	}
	etcdServiceBase64 := base64.StdEncoding.EncodeToString([]byte(etcdService))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/etcd.service\"", etcdServiceBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to generate etcd service")
	}

	etcdEnv, err := tmpl.GenerateEtcdEnv(mgr, node, mgr.Runner.Index)
	if err != nil {
		return err
	}
	etcdEnvBase64 := base64.StdEncoding.EncodeToString([]byte(etcdEnv))
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/etcd.env\"", etcdEnvBase64))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to generate etcd env")
	}

	etcdBin, err := tmpl.GenerateEtcdBinary(mgr, mgr.Runner.Index)
	if err != nil {
		return err
	}
	etcdBinBase64 := base64.StdEncoding.EncodeToString([]byte(etcdBin))
	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /usr/local/bin/etcd && chmod +x /usr/local/bin/etcd\"", etcdBinBase64))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "failed to generate etcd bin")
	}

	getEtcdCtlCmd := fmt.Sprintf("docker run --rm -v /usr/local/bin:/systembindir %s /bin/cp /usr/local/bin/etcdctl /systembindir/etcdctl", images.GetImage(mgr, "etcd").ImageName())
	_, err4 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", getEtcdCtlCmd))
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err4), "failed to get etcdctl")
	}

	_, err5 := mgr.Runner.RunCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart etcd\"")
	if err5 != nil {
		return errors.Wrap(errors.WithStack(err5), "failed to start etcd")
	}

	addrList := []string{}
	for _, host := range mgr.EtcdNodes {
		addrList = append(addrList, fmt.Sprintf("https://%s:2379", host.InternalAddress))
	}
	checkHealthCmd := fmt.Sprintf("sudo -E /bin/sh -c \"export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --endpoints=%s cluster-health | grep -q 'cluster is healthy'\"", node.Name, node.Name, etcdBinDir, strings.Join(addrList, ","))
	if mgr.Runner.Index == 0 {
		for i := 20; i > 0; i-- {
			_, err := mgr.Runner.RunCmd(checkHealthCmd)
			if err != nil {
				fmt.Println("Waiting for etcd to start")
				if i == 1 {
					return errors.Wrap(errors.WithStack(err), "failed to start etcd")
				}
			} else {
				break
			}
			time.Sleep(time.Second * 5)
		}
	}
	//else {
	//	checkMemberCmd := fmt.Sprintf("export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --no-sync --endpoints=%s member list | grep -q %s", node.HostName, node.HostName, etcdBinDir, strings.Join(addrList, ","), fmt.Sprintf("https://%s:2379", node.InternalAddress))
	//	_, err := mgr.Runner.RunCmd(checkMemberCmd)
	//	if err != nil {
	//		joinMemberCmd := fmt.Sprintf("export ETCDCTL_API=2;export ETCDCTL_CERT_FILE='/etc/ssl/etcd/ssl/admin-%s.pem';export ETCDCTL_KEY_FILE='/etc/ssl/etcd/ssl/admin-%s-key.pem';export ETCDCTL_CA_FILE='/etc/ssl/etcd/ssl/ca.pem';%s/etcdctl --endpoints=%s member add %s %s", node.HostName, node.HostName, etcdBinDir, strings.Join(addrList, ","), fmt.Sprintf("etcd%d", mgr.Runner.Index+1), fmt.Sprintf("https://%s:2380", node.InternalAddress))
	//		_, err := mgr.Runner.RunCmd(joinMemberCmd)
	//		if err != nil {
	//			fmt.Println("failed to add etcd member")
	//		}
	//	}
	//}

	for i := 20; i > 0; i-- {
		_, err := mgr.Runner.RunCmd(checkHealthCmd)
		if err != nil {
			fmt.Println("Waiting for etcd to start")
			if i == 1 {
				return errors.Wrap(errors.WithStack(err), "failed to start etcd")
			}
		} else {
			break
		}
		time.Sleep(time.Second * 5)
	}

	reloadEtcdEnvCmd := "sed -i '/ETCD_INITIAL_CLUSTER_STATE/s/\\:.*/\\: existing/g' /etc/etcd.env && systemctl daemon-reload && systemctl restart etcd"
	_, err6 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", reloadEtcdEnvCmd))
	if err6 != nil {
		return errors.Wrap(errors.WithStack(err6), "failed to reload etcd env")
	}

	for i := 20; i > 0; i-- {
		_, err := mgr.Runner.RunCmd(checkHealthCmd)
		if err != nil {
			fmt.Println("Waiting for etcd to start")
			if i == 1 {
				return errors.Wrap(errors.WithStack(err), "failed to start etcd")
			}
		} else {
			break
		}
		time.Sleep(time.Second * 5)
	}

	return nil
}
