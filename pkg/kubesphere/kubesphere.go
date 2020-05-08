package kubesphere

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	kubekeyapi "github.com/kubesphere/kubekey/pkg/apis/kubekey/v1alpha1"
	"github.com/kubesphere/kubekey/pkg/util/manager"
	"github.com/kubesphere/kubekey/pkg/util/ssh"
	"github.com/pkg/errors"
)

func DeployKubeSphere(mgr *manager.Manager) error {
	mgr.Logger.Infoln("Deploy KubeSphere")

	return mgr.RunTaskOnMasterNodes(deployKubeSphere, true)
}

func deployKubeSphere(mgr *manager.Manager, node *kubekeyapi.HostCfg, conn ssh.Connection) error {
	if mgr.Runner.Index == 0 {
		//mgr.Runner.RunCmd("sudo -E /bin/sh -c \"mkdir -p /etc/kubernetes/addons\" && /usr/local/bin/helm repo add kubesphere https://charts.kubesphere.io/qingcloud")
		out, _ := json.MarshalIndent(mgr.Cluster, "", "  ")
		fmt.Println(string(out))
		if mgr.Cluster.KubeSphere.Console.Port != 0 {
			if err := DeployKubeSphereStep(mgr); err != nil {
				return err
			}
		}
	}

	return nil
}

func DeployKubeSphereStep(mgr *manager.Manager) error {
	kubesphereYaml, err := GenerateKubeSphereYaml(mgr)
	if err != nil {
		return err
	}
	kubesphereYamlBase64 := base64.StdEncoding.EncodeToString([]byte(kubesphereYaml))
	_, err1 := mgr.Runner.RunCmd(fmt.Sprintf("sudo -E /bin/sh -c \"echo %s | base64 -d > /etc/kubernetes/addons/kubesphere.yaml\"", kubesphereYamlBase64))
	if err1 != nil {
		return errors.Wrap(errors.WithStack(err1), "failed to generate KubeSphere manifests")
	}
	_, err2 := mgr.Runner.RunCmd("/usr/local/bin/kubectl apply -f /etc/kubernetes/addons/kubesphere.yaml")
	if err2 != nil {
		return errors.Wrap(errors.WithStack(err2), "failed to deploy kubesphere.yaml")
	}

	return nil
}
