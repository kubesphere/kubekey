package cert

import (
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strings"
)

var kubeadmList = []string{
	"cd /etc/kubernetes",
	"/usr/local/bin/kubeadm alpha certs renew apiserver",
	"/usr/local/bin/kubeadm alpha certs renew apiserver-kubelet-client",
	"/usr/local/bin/kubeadm alpha certs renew front-proxy-client",
	"/usr/local/bin/kubeadm alpha certs renew admin.conf",
	"/usr/local/bin/kubeadm alpha certs renew controller-manager.conf",
	"/usr/local/bin/kubeadm alpha certs renew scheduler.conf",
}

var restartList = []string{
	"docker ps -af name=k8s_kube-apiserver* -q | xargs --no-run-if-empty docker rm -f",
	"docker ps -af name=k8s_kube-scheduler* -q | xargs --no-run-if-empty docker rm -f",
	"docker ps -af name=k8s_kube-controller-manager* -q | xargs --no-run-if-empty docker rm -f",
	"systemctl restart kubelet",
}

func ListCluster(clusterCfgFile string, logger *log.Logger, verbose bool) error {
	cfg, objName, err := config.ParseClusterCfg(clusterCfgFile, "", "", false, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}
	return Execute(executor.NewExecutor(&cfg.Spec, objName, logger, "", verbose, false, true, false, false, nil))

}
func RenewClusterCerts(clusterCfgFile string, logger *log.Logger, verbose bool) error {
	cfg, objName, err := config.ParseClusterCfg(clusterCfgFile, "", "", false, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}
	return ExecuteRenew(executor.NewExecutor(&cfg.Spec, objName, logger, "", verbose, false, true, false, false, nil))

}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}
func ExecuteRenew(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecRenewTasks(mgr)
}

func ExecTasks(mgr *manager.Manager) error {
	listTasks := []manager.Task{
		{Task: ListClusterCerts, ErrMsg: "Failed to list cluster certs."},
	}
	for _, step := range listTasks {
		if err := step.Run(mgr); err != nil {
			errors.Wrap(err, step.ErrMsg)
		}
	}
	mgr.Logger.Infoln("Successful.")
	return nil
}
func ExecRenewTasks(mgr *manager.Manager) error {
	renewTasks := []manager.Task{
		{Task: RenewClusterCert, ErrMsg: "Failed to renew cluster certs."},
	}
	for _, step := range renewTasks {
		if err := step.Run(mgr); err != nil {
			errors.Wrap(err, step.ErrMsg)
		}
	}
	mgr.Logger.Infoln("Successful.")
	return nil
}

func ListClusterCerts(m *manager.Manager) error {
	m.Logger.Infoln("Listing cluster certs ...")
	return m.RunTaskOnMasterNodes(listClusterCerts, true)
}

func listClusterCerts(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 {
		_, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cd /etc/kubernetes/pki && openssl x509 -in apiserver.crt -noout -text | grep -A 2  Validity\"", 1, true)
		if err != nil {
			return errors.Wrap(err, "Failed to get cluster certs")
		}
	}
	return nil
}

func RenewClusterCert(m *manager.Manager) error {
	m.Logger.Infoln("Renewing cluster certs ...")
	return m.RunTaskOnMasterNodes(renewClusterCerts, false)
}

func renewClusterCerts(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", strings.Join(kubeadmList, " && ")), 5, true, "printCmd")
	if err != nil {
		return errors.Wrap(err, "Failed to kubeadm alpha certs renew...")
	}
	_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", strings.Join(restartList, " && ")), 5, false, "printCmd")
	if err1 != nil {
		return errors.Wrap(err1, "Failed to restart kube-apiserver or kube-schedule or kube-controller-manager")
	}
	return nil
}
