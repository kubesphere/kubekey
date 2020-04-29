package kubernetes

import (
	"encoding/base64"
	"fmt"
	kubekeyapi "github.com/pixiake/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/pixiake/kubekey/pkg/cluster/kubernetes/tmpl"
	"github.com/pixiake/kubekey/pkg/util/manager"
	"github.com/pixiake/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

func SyncKubeBinaries(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Syncing kube binaries")

	return mgr.RunTaskOnK8sNodes(syncKubeBinaries, true)
}

func syncKubeBinaries(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	currentDir, err1 := filepath.Abs(filepath.Dir(os.Args[0]))
	if err1 != nil {
		return errors.Wrap(err1, "faild get current dir")
	}

	filepath := fmt.Sprintf("%s/%s", currentDir, kubekeyapi.DefaultPreDir)

	kubeadm := fmt.Sprintf("kubeadm-%s", mgr.Cluster.KubeCluster.Version)
	kubelet := fmt.Sprintf("kubelet-%s", mgr.Cluster.KubeCluster.Version)
	kubectl := fmt.Sprintf("kubectl-%s", mgr.Cluster.KubeCluster.Version)
	helm := fmt.Sprintf("helm-%s", kubekeyapi.DefaultHelmVersion)
	kubecni := fmt.Sprintf("cni-plugins-linux-%s-%s.tgz", kubekeyapi.DefaultArch, kubekeyapi.DefaultCniVersion)
	binaryList := []string{kubeadm, kubelet, kubectl, helm, kubecni}

	for _, binary := range binaryList {
		err2 := mgr.Runner.ScpFile(fmt.Sprintf("%s/%s", filepath, binary), fmt.Sprintf("%s/%s", "/tmp/kubekey", binary))
		if err2 != nil {
			return errors.Wrap(errors.WithStack(err2), fmt.Sprintf("failed to sync binarys"))
		}
	}

	cmdlist := []string{}

	for _, binary := range binaryList {
		if strings.Contains(binary, "cni-plugins-linux") {
			cmdlist = append(cmdlist, fmt.Sprintf("mkdir -p /opt/cni/bin && tar -zxf %s/%s -C /opt/cni/bin", "/tmp/kubekey", binary))
		} else {
			cmdlist = append(cmdlist, fmt.Sprintf("cp -f /tmp/kubekey/%s /usr/local/bin/%s && chmod +x /usr/local/bin/%s", binary, strings.Split(binary, "-")[0], strings.Split(binary, "-")[0]))
		}
	}
	cmd := strings.Join(cmdlist, " && ")
	_, err3 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"%s\"", cmd))
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), fmt.Sprintf("failed to create kubelet link"))
	}
	return nil
}

func ConfigureKubeletService(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Configure kubelet service")

	return mgr.RunTaskOnAllNodes(setKubelet, true)
}

func setKubelet(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	kubeletService, err1 := tmpl.GenerateKubeletService(mgr.Cluster)
	if err1 != nil {
		return err1
	}
	kubeletServiceBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletService))
	_, err2 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/systemd/system/kubelet.service && systemctl enable kubelet && ln -snf /usr/local/bin/kubelet /usr/bin/kubelet\"", kubeletServiceBase64))
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to generate kubelet service")
	}

	kubeletEnv, err3 := tmpl.GenerateKubeletEnv(mgr.Cluster)
	if err3 != nil {
		return err3
	}
	kubeletEnvBase64 := base64.StdEncoding.EncodeToString([]byte(kubeletEnv))
	_, err4 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"mkdir -p /etc/systemd/system/kubelet.service.d && echo %s | base64 -d > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf\"", kubeletEnvBase64))
	if err4 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to generate kubelet env")
	}

	return nil
}
