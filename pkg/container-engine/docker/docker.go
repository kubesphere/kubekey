package docker

import (
	"encoding/base64"
	"fmt"
	"github.com/lithammer/dedent"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/util"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

var (
	DockerConfigTempl = template.Must(template.New("DockerConfig").Parse(
		dedent.Dedent(`{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  "exec-opts": ["native.cgroupdriver=systemd"],
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}]
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}]
  {{- end}}
}
    `)))
)

func GenerateDockerConfig(mgr *manager.Manager) (string, error) {
	var Mirrors, InsecureRegistries string
	if mgr.Cluster.Registry.RegistryMirrors != nil {
		mirrors := []string{}
		for _, mirror := range mgr.Cluster.Registry.RegistryMirrors {
			mirrors = append(mirrors, fmt.Sprintf("\"%s\"", mirror))
		}
		Mirrors = strings.Join(mirrors, ", ")
	}
	if mgr.Cluster.Registry.InsecureRegistries != nil {
		registries := []string{}
		for _, repostry := range mgr.Cluster.Registry.InsecureRegistries {
			registries = append(registries, fmt.Sprintf("\"%s\"", repostry))
		}
		InsecureRegistries = strings.Join(registries, ", ")
	}
	return util.Render(DockerConfigTempl, util.Data{
		"Mirrors":            Mirrors,
		"InsecureRegistries": InsecureRegistries,
	})
}

func InstallerDocker(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Installing docker……")

	return mgr.RunTaskOnAllNodes(installDockerOnNode, true)
}

func installDockerOnNode(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	cmd := "sudo sh -c \"[ -z $(which docker) ] && curl https://raw.githubusercontent.com/pixiake/kubekey/master/scripts/docker-install.sh | sh ; systemctl enable docker\""
	_, err := mgr.Runner.RunCmd(cmd)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to install docker")
	}
	dockerConfig, err := GenerateDockerConfig(mgr)
	if err != nil {
		return err
	}
	dockerConfigBase64 := base64.StdEncoding.EncodeToString([]byte(dockerConfig))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/docker/daemon.json && systemctl reload docker\"", dockerConfigBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to add docker config")
	}

	return nil
}
