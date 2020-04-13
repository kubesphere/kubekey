package docker

import (
	kubekeyapi "github.com/pixiake/kubekey/apis/v1alpha1"
	"github.com/pixiake/kubekey/util/dialer/ssh"
	"github.com/pixiake/kubekey/util/state"
	"github.com/pkg/errors"
)

func InstallerDocker(s *state.State) error {
	s.Logger.Infoln("Installing docker……")

	return s.RunTaskOnAllNodes(installDockerOnNode, true)
}

func installDockerOnNode(s *state.State, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	err := installDocker(s)
	if err != nil {
		return errors.Wrap(err, "failed to install docker")
	}
	return nil
}

func installDocker(s *state.State) error {
	cmd := "sudo sh -c \"[ -z $(which docker) ] && curl https://raw.githubusercontent.com/pixiake/kubeocean/master/scripts/docker-install.sh | sh ; systemctl enable docker\""
	//cmd := "[ -z $(which docker) ] && curl https://raw.githubusercontent.com/pixiake/kubeocean/master/scripts/docker-install.sh | sh ; systemctl enable docker"
	_, _, err := s.Runner.RunRaw(cmd)

	return errors.WithStack(err)
}
