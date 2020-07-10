package kubernetes

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/kubesphere/kubekey/pkg/util"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func UpgradeKubeMasters(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Upgrading kube masters")
	return mgr.RunTaskOnMasterNodes(upgradeKubeMasters, false)
}

func upgradeKubeMasters(mgr *manager.Manager, node *kubekeyapi.HostCfg, _ ssh.Connection) error {
	version, err := mgr.Runner.ExecuteCmd("/usr/local/bin/kubelet --version", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(version, " ")[1] != mgr.Cluster.Kubernetes.Version {
		if err := SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		var kubeadmCfgBase64 string
		if util.IsExist(fmt.Sprintf("%s/kubeadm-config.yaml", mgr.WorkDir)) {
			output, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("cat %s/kubeadm-config.yaml | base64 --wrap=0", mgr.WorkDir)).CombinedOutput()
			if err != nil {
				fmt.Println(string(output))
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Failed to read custom kubeadm config: %s/kubeadm-config.yaml", mgr.WorkDir))
			}
			kubeadmCfgBase64 = strings.TrimSpace(string(output))
		} else {
			kubeadmCfg, err := tmpl.GenerateKubeadmCfg(mgr)
			if err != nil {
				return err
			}
			kubeadmCfgBase64 = base64.StdEncoding.EncodeToString([]byte(kubeadmCfg))
		}

		_, err1 := mgr.Runner.ExecuteCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes && echo %s | base64 -d > /etc/kubernetes/kubeadm-config.yaml\"", kubeadmCfgBase64), 1, false)
		if err1 != nil {
			return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
		}

		_, err2 := mgr.Runner.ExecuteCmd(fmt.Sprintf(
			"sudo -E /bin/sh -c \"/usr/local/bin/kubeadm upgrade apply -y %s --config=/etc/kubernetes/kubeadm-config.yaml "+
				"--ignore-preflight-errors=all --allow-experimental-upgrades --allow-release-candidate-upgrades --etcd-upgrade=false --certificate-renewal=true --force\"",
			mgr.Cluster.Kubernetes.Version),
			3, false)
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), fmt.Sprintf("Failed to upgrade master: %s", node.Name))
		}

		if err := SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}
	}

	return nil
}

func UpgradeKubeWorkers(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Upgrading kube workers")
	return mgr.RunTaskOnWorkerNodes(upgradeKubeWorkers, false)
}

func upgradeKubeWorkers(mgr *manager.Manager, node *kubekeyapi.HostCfg, _ ssh.Connection) error {
	version, err := mgr.Runner.ExecuteCmd("/usr/local/bin/kubelet --version", 3, false)
	if err != nil {
		return errors.Wrap(err, "Failed to get current kubelet version")
	}
	if strings.Split(version, " ")[1] != mgr.Cluster.Kubernetes.Version {

		if err := SyncKubeBinaries(mgr, node); err != nil {
			return err
		}

		if err := SetKubelet(mgr, node); err != nil {
			return err
		}

		if _, err := mgr.Runner.ExecuteCmd("sudo -E /bin/sh -c \"systemctl daemon-reload && systemctl restart kubelet\"", 2, true); err != nil {
			return err
		}
	}

	return nil
}
