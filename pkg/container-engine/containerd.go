package container_engine

import (
	"encoding/base64"
	"fmt"
	kubekeyapiv1alpha1 "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
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
)

func generateContainerdService() (string, error) {
	return util.Render(ContainerdServiceTempl, util.Data{})
}

func installContainerdOnNode(mgr *manager.Manager, node *kubekeyapiv1alpha1.HostCfg) error {
	output, _ := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"if [ -z $(which containerd) ] || [ ! -e /run/containerd/containerd.sock ]; then echo 'Container Runtime will be installed'; fi\"", 0, false)
	if strings.Contains(strings.TrimSpace(output), "Container Runtime will be installed") {
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
		// TODO
	}

	// Start Containerd
	_, err = mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl enable containerd && systemctl start containerd\"", 0, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to start containerd"))
	}

	return nil
}
