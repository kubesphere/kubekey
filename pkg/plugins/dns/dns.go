package dns

import (
	"encoding/base64"
	"fmt"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/pkg/errors"
)

func OverrideCorednsService(mgr *manager.Manager) error {
	corednsSvc, err := GenerateCorednsService(mgr)
	if err != nil {
		return err
	}
	corednsSvcgBase64 := base64.StdEncoding.EncodeToString([]byte(corednsSvc))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/coredns-svc.yaml\"", corednsSvcgBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate kubeadm config")
	}
	deleteKubednsSvcCmd := "/usr/local/bin/kubectl delete -n kube-system svc kube-dns"
	_, err2 := mgr.Runner.RunCmd(deleteKubednsSvcCmd)
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to delete kubeadm Kube-DNS service")
	}
	_, err3 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/coredns-svc.yaml")
	if err3 != nil {
		return errors.Wrap(errors.WithStack(err3), "Failed to create coredns service")
	}
	return nil
}

func DeployNodelocaldns(mgr *manager.Manager) error {
	nodelocaldns, err := GenerateNodelocaldnsService(mgr)
	if err != nil {
		return err
	}
	nodelocaldnsBase64 := base64.StdEncoding.EncodeToString([]byte(nodelocaldns))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/nodelocaldns.yaml\"", nodelocaldnsBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "Failed to generate nodelocaldns manifests")
	}
	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/nodelocaldns.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "Failed to create nodelocaldns")
	}
	return nil
}
