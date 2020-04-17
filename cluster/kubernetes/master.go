package kubernetes

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/cluster/kubernetes/tmpl"
	"github.com/pixiake/kubekey/util/manager"
	"github.com/pixiake/kubekey/util/ssh"
	"github.com/pkg/errors"
)

func InitKubernetesCluster(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Init kubernetes cluster")

	return mgr.RunTaskOnMasterNodes(initKubernetesCluster, true)
}

func initKubernetesCluster(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		kubeadmCfg, err := tmpl.GenerateKubeadmCfg(mgr)
		if err != nil {
			return err
		}
		kubeadmCfgBase64 := base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
		_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64))
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "failed to generate kubeadm config")
		}
	}

	return nil
}
