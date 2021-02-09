package rook

import (
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

func DeployRookCeph(mgr *manager.Manager) error {

	mgr.Logger.Infoln("Deploying KubeSphere ...")
	if err := mgr.RunTaskOnMasterNodes(deployRookCeph, true); err != nil {
		return err
	}

	return nil
}

func deployRookCeph(mgr *manager.Manager, _ *kubekeyapiv1alpha1.HostCfg) error {

	for filename, cephyaml := range cephmaps {
		Yml := base64.StdEncoding.EncodeToString([]byte(cephyaml))
		_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("echo %s | base64 -d > %s/%s && chmod +x %s/%s", Yml, mgr.WorkDir, filename, mgr.WorkDir, filename), 1, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate %s", filename))
		}
		if _, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"/usr/local/bin/kubectl apply -f %s/%s\"", mgr.WorkDir, filename), 1, false); err1 != nil {
			return errors.Wrap(errors.WithStack(err1), fmt.Sprintf("Failed to create yaml from %s", filename))
		}
	}
	return nil
}
