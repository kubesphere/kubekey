package docker

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/lithammer/dedent"
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
  {{- if .Mirrors }}
  "registry-mirrors": [{{ .Mirrors }}],
  {{- end}}
  {{- if .InsecureRegistries }}
  "insecure-registries": [{{ .InsecureRegistries }}],
  {{- end}}
  "exec-opts": ["native.cgroupdriver=systemd"]
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
	//cmd := "sudo sh -c \"if [ -z $(which docker) ]; then curl https://kubernetes.pek3b.qingstor.com/tools/kubekey/docker-install.sh | sh && systemctl enable docker; fi\""
	//out, err := mgr.Runner.RunCmd(cmd)
	//if err != nil {
	//	fmt.Println(out)
	//	return errors.Wrap(errors.WithStack(err), "failed to install docker")
	//}
	dockerConfig, err := GenerateDockerConfig(mgr)
	if err != nil {
		return err
	}
	dockerConfigBase64 := base64.StdEncoding.EncodeToString([]byte(dockerConfig))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -z $(which docker) ]; then curl https://kubernetes.pek3b.qingstor.com/tools/kubekey/docker-install.sh | sh && systemctl enable docker && echo %s | base64 -d > /etc/docker/daemon.json && systemctl reload docker; fi\"", dockerConfigBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to install docker")
	}

	return nil
}
