package cert

import (
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/config"
	"github.com/kubesphere/kubekey/pkg/util/executor"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func ListCluster(clusterCfgFile string, logger *log.Logger, verbose bool) error {
	cfg, objName, err := config.ParseClusterCfg(clusterCfgFile, "", "", false, logger)
	if err != nil {
		return errors.Wrap(err, "Failed to download cluster config")
	}
	return Execute(executor.NewExecutor(&cfg.Spec, objName, logger, "", verbose, false, true, false, false, nil))

}

func Execute(executor *executor.Executor) error {
	mgr, err := executor.CreateManager()
	if err != nil {
		return err
	}
	return ExecTasks(mgr)
}

func ExecTasks(mgr *manager.Manager) error {
	listTasks := []manager.Task{
		{Task: ListClusterCert, ErrMsg: "Failed to list cluster cert."},
	}
	for _, step := range listTasks {
		if err := step.Run(mgr); err != nil {
			errors.Wrap(err, step.ErrMsg)
		}
	}
	mgr.Logger.Infoln("Successful.")
	return nil
}

func ListClusterCert(m *manager.Manager) error {
	m.Logger.Infoln("Listing cluster cert ...")
	return m.RunTaskOnMasterNodes(listClusterCert, true)
}

func listClusterCert(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {
	if mgr.Runner.Index == 0 {
		_, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"cd /etc/kubernetes/pki && openssl x509 -in apiserver.crt -noout -text | grep -A 2  Validity\"", 1, true)
		if err != nil {
			return errors.Wrap(err, "Failed to get cluster cert")
		}
	}
	return nil
}
