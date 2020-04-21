package docker

import (
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func InstallerDocker(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Installing docker……")

	return mgr.RunTaskOnAllNodes(installDockerOnNode, true)
}

func installDockerOnNode(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	cmd := "sudo sh -c \"[ -z $(which docker) ] && curl https://raw.githubusercontent.com/pixiake/kubeocean/master/scripts/docker-install.sh | sh ; systemctl enable docker\""
	_, err := mgr.Runner.RunCmd(cmd)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install docker")
	}
	return nil
}
