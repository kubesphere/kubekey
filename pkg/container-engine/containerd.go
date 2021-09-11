package container_engine

import (
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/kubernetes/preinstall"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	ContainerdServiceTempl = template.Must(template.New("containerdService").Parse(
		dedent.Dedent(`[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/bin/containerd

Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=1048576
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
    `)))

	ContainerdConfigTempl = template.Must(template.New("containerdConfig").Parse(
		dedent.Dedent(`version = 2
root = "/var/lib/containerd"
state = "/run/containerd"

[grpc]
  address = "/run/containerd/containerd.sock"
  uid = 0
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216

[ttrpc]
  address = ""
  uid = 0
  gid = 0

[debug]
  address = ""
  uid = 0
  gid = 0
  level = ""

[metrics]
  address = ""
  grpc_histogram = false

[cgroup]
  path = ""

[timeouts]
  "io.containerd.timeout.shim.cleanup" = "5s"
  "io.containerd.timeout.shim.load" = "5s"
  "io.containerd.timeout.shim.shutdown" = "3s"
  "io.containerd.timeout.task.state" = "2s"

[plugins]
  [plugins."io.containerd.grpc.v1.cri"]
    sandbox_image = "{{ .SandBoxImage }}"
    [plugins."io.containerd.grpc.v1.cri".cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
      max_conf_num = 1
      conf_template = ""
    [plugins."io.containerd.grpc.v1.cri".registry]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
        {{- if .Mirrors }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = [{{ .Mirrors }}, "https://registry-1.docker.io"]
        {{ else }}
        [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
        {{- end}}
    `)))

	CrictlConfigTempl = template.Must(template.New("crictlConfig").Parse(
		dedent.Dedent(`runtime-endpoint: {{ .Endpoint }}
image-endpoint: {{ .Endpoint }}
timeout: 5
debug: false
pull-image-on-create: false
    `)))
)

func generateContainerdService() (string, error) {
	return util.Render(ContainerdServiceTempl, util.Data{})
}

func generateContainerdConfig(mgr *manager.Manager) (string, error) {
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
	return util.Render(ContainerdConfigTempl, util.Data{
		"Mirrors":            Mirrors,
		"InsecureRegistries": InsecureRegistries,
		"SandBoxImage":       preinstall.GetImage(mgr, "pause").ImageName(),
	})
}

func generateCrictlConfig() (string, error) {
	return util.Render(CrictlConfigTempl, util.Data{
		"Endpoint": kubekeyapiv1alpha1.DefaultContainerdEndpoint,
	})
}

func installContainerdOnNode(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	if !manager.ExistNode(mgr, node) {
		checkCrictl, _ := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"if [ -z $(which crictl) ]; then echo 'Crictl will be installed'; fi\"", 0, false)
		if strings.Contains(strings.TrimSpace(checkCrictl), "Crictl will be installed") {
			err := syncCrictlBinarie(mgr, node)
			if err != nil {
				return err
			}
		}

		checkContainerd, _ := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; then echo 'Container Runtime will be installed'; fi\"", 0, false)
		if strings.Contains(strings.TrimSpace(checkContainerd), "Container Runtime will be installed") {
			// Installation and configuration of Containerd multiplex Docker.
			err := syncDockerBinaries(mgr, node)
			if err != nil {
				return err
			}
			err = setContainerd(mgr)
			if err != nil {
				return err
			}
			err = setDocker(mgr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func syncCrictlBinarie(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	tmpDir := kubekeyapiv1alpha1.DefaultTmpDir
	_, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"if [ -d %s ]; then rm -rf %s ;fi\" && mkdir -p %s", tmpDir, tmpDir, tmpDir), 1, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to create tmp dir")
	}

	currentDir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
	if err1 != nil {
		return errors.Wrap(err1, "Failed to get current dir")
	}

	filesDir := fmt.Sprintf("%s/%s/%s/%s", currentDir, kubekeyapiv1alpha1.DefaultPreDir, mgr.Cluster.Kubernetes.Version, node.Arch)

	crictl := fmt.Sprintf("crictl-%s-linux-%s.tar.gz", kubekeyapiv1alpha1.DefaultCrictlVersion, node.Arch)

	if err := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filesDir, crictl), fmt.Sprintf("%s/%s", "/tmp/kubekey", crictl)); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to sync binaries"))
	}

	if _, err := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /usr/bin && tar -zxf %s/%s -C /usr/bin \"", "/tmp/kubekey", crictl), 2, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to install crictl binary"))
	}

	return nil
}

func setContainerd(mgr *manager.Manager) error {
	// Generate systemd service for containerd
	containerdService, err := generateContainerdService()
	if err != nil {
		return err
	}
	containerdServiceBase64 := base64.StdEncoding.EncodeToString([]byte(containerdService))
	_, err = mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/containerd.service\"", containerdServiceBase64), 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate containerd's service"))
	}

	if strings.TrimSpace(mgr.Cluster.Kubernetes.ContainerManager) == "containerd" {
		// Generate configuration file for containerd
		containerdConfig, err := generateContainerdConfig(mgr)
		if err != nil {
			return err
		}
		containerdConfigBase64 := base64.StdEncoding.EncodeToString([]byte(containerdConfig))
		_, err = mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/containerd && echo %s | base64 -d > /etc/containerd/config.toml\"", containerdConfigBase64), 0, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate containerd's configuration"))
		}

		// Generate configuration file for crictl
		crictlConfig, err := generateCrictlConfig()
		if err != nil {
			return err
		}
		crictlConfigBase64 := base64.StdEncoding.EncodeToString([]byte(crictlConfig))
		_, err = mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/crictl.yaml\"", crictlConfigBase64), 0, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to generate crictl's configuration"))
		}
	}

	// Start Containerd
	_, err = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl enable containerd && systemctl start containerd\"", 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to start containerd"))
	}

	return nil
}
